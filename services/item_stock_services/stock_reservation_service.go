package item_stock_services

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// getAvailableSnapshot is physical stock (summed across every bin) minus approved
// reservations for one item - the same formula as GetAvailableStock below, just scoped
// to a single item and callable inside an existing tx (GetAvailableStock always reads
// via initializers.DB, which wouldn't see this transaction's own uncommitted writes).
func (s *ItemStockService) getAvailableSnapshot(tx *gorm.DB, itemId uint) (int, error) {
	var physical int
	if err := tx.Raw(`SELECT ISNULL(SUM(stock_qty), 0) FROM tbl_inv_item_stocks WHERE item_id = ?`, itemId).Scan(&physical).Error; err != nil {
		return 0, err
	}

	// Only an approved reservation holds stock (10.4.3, 14.22).
	var reserved int
	if err := tx.Raw(`SELECT ISNULL(SUM(qty), 0) FROM tbl_inv_stock_reservations WHERE item_id = ? AND status = ?`, itemId, inventory_models.ReservationStatusApproved).Scan(&reserved).Error; err != nil {
		return 0, err
	}

	return physical - reserved, nil
}

// logReservationLedger records a reservation/release as its own row in
// tbl_inv_stock_transactions - the same ledger tr_inv_item_stocks_ledger writes to for
// every physical movement, but written directly here instead, since nothing in
// tbl_inv_item_stocks itself changes when stock is only reserved (that trigger would
// never fire). Direction is "RESERVE"/"RELEASE" rather than the trigger's "IN"/"OUT", so
// anyone reading the ledger can tell these apart from a real physical movement.
// QtyBefore/QtyAfter/QtyChange track *available* stock (physical minus reservations),
// not physical stock, since that's the number a reservation actually moves. Written only
// when availability really changes: on approval, and when an approved reservation goes.
func (s *ItemStockService) logReservationLedger(tx *gorm.DB, itemId uint, direction string, qtyBefore int, qtyAfter int, qtyChange int, sourceType string, sourceId uint, quotationId uint, dbUser string, remarks string) error {
	st := sourceType
	sid := sourceId
	rm := remarks

	entry := &inventory_models.StockTransaction{
		// No single tbl_inv_item_stocks row backs a reservation (it isn't bin/warehouse
		// scoped), so there's nothing real to put here.
		RefId:         0,
		ItemId:        itemId,
		WarehouseId:   0,
		BinLocation:   "",
		DocNo:         int(quotationId),
		Direction:     direction,
		QtyBefore:     qtyBefore,
		QtyAfter:      qtyAfter,
		QtyChange:     qtyChange,
		SourceType:    &st,
		SourceId:      &sid,
		Remarks:       &rm,
		TransactionAt: time.Now(),
		DbUser:        dbUser,
	}

	return services.DbInsert(tx, entry)
}

// quoteWindow is the part of a quotation a reservation's limit depends on.
type quoteWindow struct {
	ValidityDays string
	ValidUntil   string
}

func (s *ItemStockService) loadQuoteWindow(tx *gorm.DB, quotationId uint) (quoteWindow, error) {
	var q quoteWindow
	if quotationId == 0 {
		return q, nil
	}
	err := tx.Raw(`SELECT ISNULL(validity_days, '') AS validity_days, ISNULL(valid_until, '') AS valid_until
		FROM tbl_trans_sales_quotation WHERE id = ?`, quotationId).Scan(&q).Error
	return q, err
}

// CreateStockReservation records a reservation request for a quotation line (10.4.1). A
// request moves no stock: that happens when the Warehouse Manager approves it, which is
// also when its RESERVE ledger row is written (setReservationDecision).
//
// The limit is the quote's VALID UNTIL (10.4.5), read off the quotation itself; the
// caller's expiresAt is used only when the quotation has none to give.
func (s *ItemStockService) CreateStockReservation(tx *gorm.DB, itemId uint, qty uint, sourceType string, sourceId uint, quotationId uint, expiresAt *time.Time, dbUser string) error {
	if qty == 0 {
		return nil
	}

	quote, err := s.loadQuoteWindow(tx, quotationId)
	if err != nil {
		return err
	}
	if limit := parseQuoteDateTime(quote.ValidUntil); limit != nil {
		expiresAt = limit
	}

	reservation := &inventory_models.StockReservation{
		ItemId:      itemId,
		Qty:         qty,
		SourceType:  sourceType,
		SourceId:    sourceId,
		QuotationId: quotationId,
		ReservedAt:  time.Now(),
		ExpiresAt:   expiresAt,
		Status:      inventory_models.ReservationStatusPending,
	}

	return services.DbInsert(tx, reservation)
}

// deleteReservation removes one reservation, logging a RELEASE row only when it was
// approved - a pending or declined one never held stock, so nothing returns.
func (s *ItemStockService) deleteReservation(tx *gorm.DB, r *inventory_models.StockReservation, dbUser string, remarks string) error {
	wasHolding := r.Status == inventory_models.ReservationStatusApproved

	availableBefore := 0
	if wasHolding {
		var err error
		if availableBefore, err = s.getAvailableSnapshot(tx, r.ItemId); err != nil {
			return err
		}
	}

	if err := services.DbDelete(tx, &inventory_models.StockReservation{}, map[string]interface{}{"id": r.ID}); err != nil {
		return err
	}

	if !wasHolding {
		return nil
	}
	return s.logReservationLedger(tx, r.ItemId, "RELEASE", availableBefore, availableBefore+int(r.Qty), int(r.Qty), r.SourceType, r.SourceId, r.QuotationId, dbUser, remarks)
}

// ReleaseStockReservation removes every reservation for one source line (a quotation line
// being deleted, or sales unticking RESERVE). A line with none is not an error.
func (s *ItemStockService) ReleaseStockReservation(tx *gorm.DB, sourceType string, sourceId uint, dbUser string) error {
	var rows []inventory_models.StockReservation
	if err := tx.Where("source_type = ? AND source_id = ?", sourceType, sourceId).Find(&rows).Error; err != nil {
		return err
	}

	for i := range rows {
		remarks := fmt.Sprintf("Released from sales quotation #%d", rows[i].QuotationId)
		if err := s.deleteReservation(tx, &rows[i], dbUser, remarks); err != nil {
			return err
		}
	}

	return nil
}

// CreateStockReservationByRef is the standalone entry point for sales ticking "RESERVE" on
// the stock-check modal - the counterpart to ReleaseStockReservationByRef below.
func (s *ItemStockService) CreateStockReservationByRef(itemId uint, qty uint, sourceType string, sourceId uint, quotationId uint, expiresAt *time.Time, dbUser string) (int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := s.CreateStockReservation(tx, itemId, qty, sourceType, sourceId, quotationId, expiresAt, dbUser); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed creating stock reservation")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// GetReservation looks up whether a specific quotation line currently has a reservation -
// the stock-check modal needs this to know whether to draw its RESERVE checkbox as checked
// when it's first opened. Returns (nil, nil) if there's no reservation for that ref.
func (s *ItemStockService) GetReservation(sourceType string, sourceId uint) (*inventory_models.StockReservation, error) {
	var reservation inventory_models.StockReservation
	err := initializers.DB.Where("source_type = ? AND source_id = ?", sourceType, sourceId).First(&reservation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &reservation, nil
}

// SyncReservationQty keeps an existing reservation's qty in line with its quotation line's
// edited QTY; it never creates one. Spec 10.4.5: a quantity change is a new request, so it
// goes back to the Warehouse Manager (10.4.2), and its window restarts from the change date
// with the quote's VALID UNTIL moving to match.
func (s *ItemStockService) SyncReservationQty(tx *gorm.DB, sourceType string, sourceId uint, qty uint) error {
	var rows []inventory_models.StockReservation
	if err := tx.Where("source_type = ? AND source_id = ? AND status <> ?", sourceType, sourceId, inventory_models.ReservationStatusRejected).
		Order("id desc").Limit(1).Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	existing := rows[0]

	// Zero means QTY was not part of this save - DbUpdate skips zero values the same way. A
	// line that is really gone is released by its own delete path.
	if qty == 0 || existing.Qty == qty {
		return nil
	}

	if existing.Status == inventory_models.ReservationStatusApproved {
		// The old quantity stops holding stock until the new request is approved.
		availableBefore, err := s.getAvailableSnapshot(tx, existing.ItemId)
		if err != nil {
			return err
		}
		remarks := fmt.Sprintf("Quantity changed on sales quotation #%d - awaiting the Warehouse Manager again", existing.QuotationId)
		if err := s.logReservationLedger(tx, existing.ItemId, "RELEASE", availableBefore, availableBefore+int(existing.Qty), int(existing.Qty), existing.SourceType, existing.SourceId, existing.QuotationId, "system", remarks); err != nil {
			return err
		}
	}

	if err := tx.Model(&inventory_models.StockReservation{}).Where("id = ?", existing.ID).
		Updates(map[string]interface{}{
			"qty":              qty,
			"status":           inventory_models.ReservationStatusPending,
			"approved_by":      nil,
			"approved_at":      nil,
			"limit_reached_at": nil,
		}).Error; err != nil {
		return err
	}

	_, err := s.restartQuoteWindow(tx, existing.QuotationId, existing.ID, time.Now())
	return err
}

// restartQuoteWindow starts a fresh reservation window at `from` (10.4.5): the quote's VALID
// UNTIL becomes from + VALIDITY (DAYS), and every reservation on that quote takes the same
// limit, so "the two never diverge". A reservation with no quotation gets the window alone.
func (s *ItemStockService) restartQuoteWindow(tx *gorm.DB, quotationId uint, reservationId uint, from time.Time) (time.Time, error) {
	quote, err := s.loadQuoteWindow(tx, quotationId)
	if err != nil {
		return time.Time{}, err
	}
	end := reservationWindowEnd(from, quote.ValidityDays)
	clearLimit := map[string]interface{}{"expires_at": end, "limit_reached_at": nil}

	if quotationId == 0 {
		return end, tx.Model(&inventory_models.StockReservation{}).Where("id = ?", reservationId).Updates(clearLimit).Error
	}

	if err := tx.Exec(`UPDATE tbl_trans_sales_quotation SET valid_until = ? WHERE id = ?`, end.Format(quoteDateTimeFormat), quotationId).Error; err != nil {
		return end, fmt.Errorf("failed moving the quote's valid until: %w", err)
	}
	if err := services.InvalidateCache(services.GetKey(&models.SalesQuotation{}, nil)); err != nil {
		return end, err
	}
	if err := services.InvalidateCacheByModel(&models.SalesQuotation{}); err != nil {
		return end, err
	}

	return end, tx.Model(&inventory_models.StockReservation{}).Where("quotation_id = ?", quotationId).Updates(clearLimit).Error
}

// MoveReservationLimits carries a quote's edited VALID UNTIL onto its reservations (10.4.5:
// "extending or shortening the quote's validity moves the reservation limit with it"). A
// limit moved back into the future also clears an unanswered at-limit question.
func (s *ItemStockService) MoveReservationLimits(tx *gorm.DB, quotationId uint, validUntil string) error {
	end := parseQuoteDateTime(validUntil)
	if quotationId == 0 || end == nil {
		return nil
	}

	updates := map[string]interface{}{"expires_at": *end}
	if end.After(time.Now()) {
		updates["limit_reached_at"] = nil
	}

	return tx.Model(&inventory_models.StockReservation{}).Where("quotation_id = ?", quotationId).Updates(updates).Error
}

// ReleaseStockReservationByRef is the standalone entry point for releasing a reservation
// from the UI directly - sales unticking "RESERVE" on the stock-check modal.
// CreateStockReservation/ReleaseStockReservation above take an existing tx because they're
// called from inside the quotation create/update/delete flow; this opens and commits its
// own, the same way InsertItemStock/AdjustItemStock do.
func (s *ItemStockService) ReleaseStockReservationByRef(sourceType string, sourceId uint, dbUser string) (int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := s.ReleaseStockReservation(tx, sourceType, sourceId, dbUser); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed releasing stock reservation")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// FlagReservationsAtLimit marks every reservation that has reached its limit (10.4.5) and
// has not been flagged yet. It releases nothing and extends nothing (14.23): the reservation
// waits, still holding whatever it held, until the Warehouse Manager or the owning sales
// executive answers (AnswerReservationLimit). The flag is what puts the question to them -
// the Reservations submodule and the sales red box both read it. main.go's
// startStockReservationSweep runs this.
func (s *ItemStockService) FlagReservationsAtLimit(tx *gorm.DB) (int64, error) {
	now := time.Now()
	result := tx.Model(&inventory_models.StockReservation{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND limit_reached_at IS NULL AND status <> ?", now, inventory_models.ReservationStatusRejected).
		Update("limit_reached_at", now)
	return result.RowsAffected, result.Error
}

// GetAvailableStock returns physical stock (summed across every bin) minus approved
// reservations for that item, so a quotation screen can show what's actually free to
// promise. itemId = 0 returns every item.
func (s *ItemStockService) GetAvailableStock(itemId uint) ([]inventory_models.AvailableStockView, int, error) {
	var response []inventory_models.AvailableStockView

	query := `
		SELECT
			p.item_id,
			p.physical,
			ISNULL(r.reserved, 0) AS reserved,
			p.physical - ISNULL(r.reserved, 0) AS available
		FROM (
			SELECT item_id, SUM(stock_qty) AS physical
			FROM tbl_inv_item_stocks
			WHERE (? = 0 OR item_id = ?)
			GROUP BY item_id
		) p
		LEFT JOIN (
			SELECT item_id, SUM(qty) AS reserved
			FROM tbl_inv_stock_reservations
			WHERE (? = 0 OR item_id = ?) AND status = ?
			GROUP BY item_id
		) r ON r.item_id = p.item_id
	`

	if err := initializers.DB.Raw(query, itemId, itemId, itemId, itemId, inventory_models.ReservationStatusApproved).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting available stock")
	}

	return response, fiber.StatusOK, nil
}

// ReservationApprovalAccessCode is the tbl_position_access code that makes a Position the
// Warehouse Manager for reservations (10.4.2) - the same module-access mechanism used
// everywhere else in this app (see position_access_service.go), applied to a business
// action instead of a screen. Nothing here hardcodes a position name.
const ReservationApprovalAccessCode = "RESERVATION_APPROVAL"

// UserCanApproveReservations checks whether the given user's Position has been granted
// ReservationApprovalAccessCode.
func (s *ItemStockService) UserCanApproveReservations(userId uint) (bool, error) {
	if userId == 0 {
		return false, nil
	}

	var count int64
	err := initializers.DB.Raw(`
		SELECT COUNT(*)
		FROM tbl_position_access pa
		INNER JOIN tbl_setup_users u ON u.position_id = pa.position_id
		WHERE u.id = ? AND pa.code = ?
	`, userId, ReservationApprovalAccessCode).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// userOwnsQuote reports whether the user is the sales executive who owns the quote.
// Quotations record their owner as the creator's "Firstname Lastname" in created_by, with
// no user id behind it - the sales red box matches on the same string.
func (s *ItemStockService) userOwnsQuote(tx *gorm.DB, userId uint, quotationId uint) (bool, error) {
	if userId == 0 || quotationId == 0 {
		return false, nil
	}

	var count int64
	err := tx.Raw(`
		SELECT COUNT(*)
		FROM tbl_setup_users u
		INNER JOIN tbl_trans_sales_quotation q ON q.id = ?
		WHERE u.id = ?
		  AND LTRIM(RTRIM(ISNULL(q.created_by, ''))) <> ''
		  AND LTRIM(RTRIM(u.first_name)) + ' ' + LTRIM(RTRIM(u.last_name)) = LTRIM(RTRIM(q.created_by))
	`, quotationId, userId).Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ApproveReservation puts a reservation into effect (10.4.2): its units leave availability
// and a RESERVE row is written to the ledger.
func (s *ItemStockService) ApproveReservation(reservationId uint, approvedByUserId uint) (int, error) {
	return s.setReservationDecision(reservationId, approvedByUserId, inventory_models.ReservationStatusApproved)
}

// RejectReservation declines a reservation, or withdraws an approved one. The row stays
// (10.4.2), and anything it held returns to availability.
func (s *ItemStockService) RejectReservation(reservationId uint, rejectedByUserId uint) (int, error) {
	return s.setReservationDecision(reservationId, rejectedByUserId, inventory_models.ReservationStatusRejected)
}

func (s *ItemStockService) setReservationDecision(reservationId uint, actingUserId uint, newStatus string) (int, error) {
	canApprove, err := s.UserCanApproveReservations(actingUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to approve or decline reservations")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var reservation inventory_models.StockReservation
	if err := tx.First(&reservation, reservationId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("reservation not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading reservation")
	}

	if !canDecide(reservation.Status, newStatus) {
		return fiber.StatusConflict, fmt.Errorf("reservation is already %s", reservation.Status)
	}

	availableBefore, err := s.getAvailableSnapshot(tx, reservation.ItemId)
	if err != nil {
		return fiber.StatusInternalServerError, err
	}
	wasHolding := reservation.Status == inventory_models.ReservationStatusApproved

	if err := tx.Model(&inventory_models.StockReservation{}).Where("id = ?", reservation.ID).
		Updates(map[string]interface{}{
			"status":      newStatus,
			"approved_by": actingUserId,
			"approved_at": time.Now(),
		}).Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating reservation")
	}

	user := fmt.Sprintf("user#%d", actingUserId)
	qty := int(reservation.Qty)

	switch {
	case newStatus == inventory_models.ReservationStatusApproved:
		remarks := fmt.Sprintf("Reservation approved for sales quotation #%d", reservation.QuotationId)
		err = s.logReservationLedger(tx, reservation.ItemId, "RESERVE", availableBefore, availableBefore-qty, -qty, reservation.SourceType, reservation.SourceId, reservation.QuotationId, user, remarks)
	case wasHolding:
		remarks := fmt.Sprintf("Reservation declined for sales quotation #%d", reservation.QuotationId)
		err = s.logReservationLedger(tx, reservation.ItemId, "RELEASE", availableBefore, availableBefore+qty, qty, reservation.SourceType, reservation.SourceId, reservation.QuotationId, user, remarks)
	}
	if err != nil {
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// RemoveReservation is the Reservations submodule's ✕ (10.4.2): the row is deleted, used
// when sales advises the quote is off. Warehouse Manager only.
func (s *ItemStockService) RemoveReservation(reservationId uint, actingUserId uint) (int, error) {
	canApprove, err := s.UserCanApproveReservations(actingUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to remove reservations")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var reservation inventory_models.StockReservation
	if err := tx.First(&reservation, reservationId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("reservation not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading reservation")
	}

	remarks := fmt.Sprintf("Reservation removed for sales quotation #%d", reservation.QuotationId)
	if err := s.deleteReservation(tx, &reservation, fmt.Sprintf("user#%d", actingUserId), remarks); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed removing reservation")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// AnswerReservationLimit records the answer to a reservation's at-limit question (10.4.5).
// The Warehouse Manager (RESERVATION_APPROVAL) or the sales executive who owns the quote may
// answer, and the first answer settles it:
//   - keep on hold: a fresh window begins from the answer, the quote's VALID UNTIL moves with
//     it, and the reservation keeps its status and stock;
//   - let go: the reservation is removed and anything it held returns to TOTAL STOCK.
//
// The Warehouse Manager can still decline or remove a reservation sales chose to keep
// (10.4.2: no one in sales may override them).
func (s *ItemStockService) AnswerReservationLimit(reservationId uint, actingUserId uint, keep bool) (int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	var reservation inventory_models.StockReservation
	if err := tx.First(&reservation, reservationId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.StatusNotFound, errors.New("reservation not found")
		}
		return fiber.StatusInternalServerError, errors.New("failed loading reservation")
	}

	if reservation.LimitReachedAt == nil {
		return fiber.StatusConflict, errors.New("this reservation has not reached its limit")
	}

	allowed, err := s.UserCanApproveReservations(actingUserId)
	if err == nil && !allowed {
		allowed, err = s.userOwnsQuote(tx, actingUserId, reservation.QuotationId)
	}
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking who may answer")
	}
	if !allowed {
		return fiber.StatusForbidden, errors.New("only the Warehouse Manager or the sales executive who owns the quote can answer this")
	}

	if keep {
		if _, err := s.restartQuoteWindow(tx, reservation.QuotationId, reservation.ID, time.Now()); err != nil {
			return fiber.StatusInternalServerError, errors.New("failed starting a fresh reservation window")
		}
	} else {
		remarks := fmt.Sprintf("Let go at its limit for sales quotation #%d", reservation.QuotationId)
		if err := s.deleteReservation(tx, &reservation, fmt.Sprintf("user#%d", actingUserId), remarks); err != nil {
			return fiber.StatusInternalServerError, errors.New("failed letting the reservation go")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// reservationViewSelect is one row per reservation, joined with just enough item/quotation
// context to review it without a separate lookup per row.
const reservationViewSelect = `
		SELECT
			r.id,
			r.item_id,
			ISNULL(n.name, '') AS item_name,
			ISNULL(i.item_model, '') AS item_model,
			ISNULL(i.item_code, '') AS item_code,
			r.qty,
			r.source_type,
			r.source_id,
			r.quotation_id,
			ISNULL(q.document_no, '') AS document_no,
			ISNULL(cust.branch_name, '') AS customer_name,
			ISNULL(q.project_name, '') AS project_name,
			ISNULL(q.created_by, '') AS requested_by,
			r.reserved_at,
			r.expires_at,
			r.status,
			r.limit_reached_at
		FROM tbl_inv_stock_reservations r
		LEFT JOIN tbl_setup_item i ON i.id = r.item_id
		LEFT JOIN tbl_setup_item_name n ON n.id = i.item_name_id
		LEFT JOIN tbl_trans_sales_quotation q ON q.id = r.quotation_id
		-- q.customer_id is a tbl_bpi.id; the display name is one hop away in
		-- tbl_bpi_general (same join the GetBpiCustomer view makes). OUTER APPLY
		-- rather than a plain LEFT JOIN because a BPI can carry several branch
		-- rows - a join would fan one reservation out into several queue rows.
		OUTER APPLY (
			SELECT TOP 1 g.branch_name
			FROM tbl_bpi_general g
			WHERE g.based_id = q.customer_id
			ORDER BY CASE WHEN g.is_main = 1 THEN 0 ELSE 1 END, g.id
		) cust`

// GetReservationQueue backs the Reservations submodule (10.4.2): every reservation, with
// the ones waiting on an at-limit answer first, then pending requests, then approved and
// declined rows (a declined row stays, per 10.4.2). atLimitOnly narrows it to the at-limit
// questions, which is what the sales red box asks for.
func (s *ItemStockService) GetReservationQueue(atLimitOnly bool) ([]inventory_models.PendingReservationView, int, error) {
	var response []inventory_models.PendingReservationView

	where := ""
	if atLimitOnly {
		where = " WHERE r.limit_reached_at IS NOT NULL"
	}

	query := reservationViewSelect + where + `
		ORDER BY
			CASE WHEN r.limit_reached_at IS NOT NULL THEN 0
			     WHEN r.status = 'Pending' THEN 1
			     WHEN r.status = 'Approved' THEN 2
			     ELSE 3 END,
			r.reserved_at ASC`

	if err := initializers.DB.Raw(query).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting reservations")
	}

	return response, fiber.StatusOK, nil
}

// GetPendingReservations lists only the requests still awaiting a decision, oldest first.
func (s *ItemStockService) GetPendingReservations() ([]inventory_models.PendingReservationView, int, error) {
	var response []inventory_models.PendingReservationView

	query := reservationViewSelect + `
		WHERE r.status = ?
		ORDER BY r.reserved_at ASC`

	if err := initializers.DB.Raw(query, inventory_models.ReservationStatusPending).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting pending reservations")
	}

	return response, fiber.StatusOK, nil
}
