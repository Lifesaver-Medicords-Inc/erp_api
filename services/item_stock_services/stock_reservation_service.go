package item_stock_services

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// getAvailableSnapshot is physical stock (summed across every bin) minus current
// reservations for one item - the same formula as GetAvailableStock below, just scoped
// to a single item and callable inside an existing tx (GetAvailableStock always reads
// via initializers.DB, which wouldn't see this transaction's own uncommitted writes).
func (s *ItemStockService) getAvailableSnapshot(tx *gorm.DB, itemId uint) (int, error) {
	var physical int
	if err := tx.Raw(`SELECT ISNULL(SUM(stock_qty), 0) FROM tbl_inv_item_stocks WHERE item_id = ?`, itemId).Scan(&physical).Error; err != nil {
		return 0, err
	}

	// Pending AND Approved both still hold the stock (see the Status doc comment on
	// StockReservation) - only Rejected drops out.
	var reserved int
	if err := tx.Raw(`SELECT ISNULL(SUM(qty), 0) FROM tbl_inv_stock_reservations WHERE item_id = ? AND status <> ?`, itemId, inventory_models.ReservationStatusRejected).Scan(&reserved).Error; err != nil {
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
// not physical stock, since that's the number a reservation actually moves.
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

// CreateStockReservation places a soft hold for a quotation line. expiresAt is
// whatever the caller resolved from the parent document (e.g. a quotation's
// ValidUntil) - pass nil if it couldn't be parsed, but note that means
// ExpireStockReservations will never clean this row up on its own. Also writes a
// "RESERVE" row to the stock ledger (see logReservationLedger) so the reservation shows
// up as a deduction against available stock there, even though physical stock (and
// therefore the trigger-driven side of the ledger) never moves.
func (s *ItemStockService) CreateStockReservation(tx *gorm.DB, itemId uint, qty uint, sourceType string, sourceId uint, quotationId uint, expiresAt *time.Time, dbUser string) error {
	if qty == 0 {
		return nil
	}

	availableBefore, err := s.getAvailableSnapshot(tx, itemId)
	if err != nil {
		return err
	}

	reservation := &inventory_models.StockReservation{
		ItemId:      itemId,
		Qty:         qty,
		SourceType:  sourceType,
		SourceId:    sourceId,
		QuotationId: quotationId,
		ReservedAt:  time.Now(),
		ExpiresAt:   expiresAt,
		// Always starts out awaiting sign-off from a dispatcher/inventory manager (see
		// ApproveReservation/RejectReservation) - still counts as a soft hold against
		// available stock while pending, it just isn't authorized yet.
		Status: inventory_models.ReservationStatusPending,
	}

	if err := services.DbInsert(tx, reservation); err != nil {
		return err
	}

	availableAfter := availableBefore - int(qty)
	remarks := fmt.Sprintf("Reserved for sales quotation #%d", quotationId)
	return s.logReservationLedger(tx, itemId, "RESERVE", availableBefore, availableAfter, -int(qty), sourceType, sourceId, quotationId, dbUser, remarks)
}

// ReleaseStockReservation removes the hold for one source line (e.g. a quotation line
// being deleted, or a manager unchecking RESERVE). A row simply not existing means "not
// reserved" - there's no is_active flag to flip, and nothing to log either. Also writes
// a "RELEASE" row to the stock ledger mirroring CreateStockReservation's "RESERVE" row -
// see logReservationLedger.
func (s *ItemStockService) ReleaseStockReservation(tx *gorm.DB, sourceType string, sourceId uint, dbUser string) error {
	var existing inventory_models.StockReservation
	err := tx.Where("source_type = ? AND source_id = ?", sourceType, sourceId).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	availableBefore, err := s.getAvailableSnapshot(tx, existing.ItemId)
	if err != nil {
		return err
	}

	if err := services.DbDelete(tx, &inventory_models.StockReservation{}, map[string]interface{}{
		"source_type": sourceType,
		"source_id":   sourceId,
	}); err != nil {
		return err
	}

	availableAfter := availableBefore + int(existing.Qty)
	remarks := fmt.Sprintf("Released from sales quotation #%d", existing.QuotationId)
	return s.logReservationLedger(tx, existing.ItemId, "RELEASE", availableBefore, availableAfter, int(existing.Qty), sourceType, sourceId, existing.QuotationId, dbUser, remarks)
}

// CreateStockReservationByRef is the standalone entry point for a sales rep/manager
// manually checking "RESERVE" on the stock-check modal - the counterpart to
// ReleaseStockReservationByRef below. Reservations are no longer placed automatically
// when a quotation line is created (see quick_quotation_service.go) - this is the only
// place one gets created from the UI now.
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

// GetReservation looks up whether a specific quotation line currently has a manual
// reservation - the stock-check modal needs this to know whether to draw its RESERVE
// checkbox as checked when it's first opened, since (unlike before) nothing reserves a
// line automatically anymore. Returns (nil, nil) if there's no reservation for that ref.
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

// SyncReservationQty keeps an already-existing manual reservation's qty in line with a
// quotation line's edited QTY. It deliberately does NOT create a reservation where none
// exists - reserving is an explicit, manager-gated action from StockCheckModal, not
// something editing a quotation line should trigger on its own.
func (s *ItemStockService) SyncReservationQty(tx *gorm.DB, sourceType string, sourceId uint, qty uint) error {
	var existing inventory_models.StockReservation
	err := tx.Where("source_type = ? AND source_id = ?", sourceType, sourceId).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Nothing reserved for this line - nothing to sync.
			return nil
		}
		return err
	}

	if existing.Qty == qty {
		return nil
	}

	existing.Qty = qty
	return services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID})
}

// ReleaseStockReservationByRef is the standalone entry point for releasing a reservation
// from the UI directly - e.g. a sales manager unchecking "RESERVE" on the stock-check
// modal to free stock back to the pool before the line's own ValidUntil expiry would
// have done it automatically. CreateStockReservation/ReleaseStockReservation above take
// an existing tx because they're called from inside the quotation create/update/delete
// flow; this opens and commits its own, the same way InsertItemStock/AdjustItemStock do.
//
// Re-reserving afterward means checking RESERVE again from the UI (CreateStockReservationByRef
// above) - editing the line's qty no longer recreates a released reservation, it only
// syncs the qty of one that's still active (see SyncReservationQty).
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

// ExpireStockReservations deletes every reservation whose ExpiresAt has passed, logging
// a "RELEASE" ledger row for each one it removes (see logReservationLedger) - same as a
// manual release, just triggered by time instead of someone unchecking RESERVE. Nothing
// in this codebase calls this on its own - see initializers.StartReservationSweep for
// the periodic goroutine that does.
func (s *ItemStockService) ExpireStockReservations(tx *gorm.DB) (int64, error) {
	var expired []inventory_models.StockReservation
	if err := tx.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Find(&expired).Error; err != nil {
		return 0, err
	}

	if len(expired) == 0 {
		return 0, nil
	}

	for _, reservation := range expired {
		availableBefore, err := s.getAvailableSnapshot(tx, reservation.ItemId)
		if err != nil {
			return 0, err
		}

		if err := services.DbDelete(tx, &inventory_models.StockReservation{}, map[string]interface{}{"id": reservation.ID}); err != nil {
			return 0, err
		}

		availableAfter := availableBefore + int(reservation.Qty)
		remarks := fmt.Sprintf("Reservation expired for sales quotation #%d", reservation.QuotationId)
		if err := s.logReservationLedger(tx, reservation.ItemId, "RELEASE", availableBefore, availableAfter, int(reservation.Qty), reservation.SourceType, reservation.SourceId, reservation.QuotationId, "system", remarks); err != nil {
			return 0, err
		}
	}

	return int64(len(expired)), nil
}

// GetAvailableStock returns physical stock (summed across every bin) minus whatever's
// currently reserved for that item, so a quotation screen can show what's actually
// free to promise. itemId = 0 returns every item.
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
			WHERE (? = 0 OR item_id = ?) AND status <> ?
			GROUP BY item_id
		) r ON r.item_id = p.item_id
	`

	if err := initializers.DB.Raw(query, itemId, itemId, itemId, itemId, inventory_models.ReservationStatusRejected).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting available stock")
	}

	return response, fiber.StatusOK, nil
}

// ReservationApprovalAccessCode is the tbl_position_access code that grants a Position
// the ability to approve/reject pending stock reservations - the same module-access
// mechanism used everywhere else in this app (see position_access_service.go), just
// applied to a business action instead of a screen. Grant it to whatever Position
// actually handles this (Dispatcher, Inventory Manager, or whatever it ends up being
// called) from the normal Position Access setup screen - nothing here hardcodes a
// position name.
const ReservationApprovalAccessCode = "RESERVATION_APPROVAL"

// UserCanApproveReservations checks whether the given user's Position has been granted
// ReservationApprovalAccessCode.
func (s *ItemStockService) UserCanApproveReservations(userId uint) (bool, error) {
	// TEMP DEBUG - remove once RESERVATION_APPROVAL 403s are confirmed fixed.
	fmt.Printf("[RESV-DEBUG] UserCanApproveReservations called with userId=%d\n", userId)

	if userId == 0 {
		fmt.Println("[RESV-DEBUG] userId is 0 - short-circuiting to false without querying the DB")
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
		fmt.Printf("[RESV-DEBUG] query error: %v\n", err)
		return false, err
	}

	fmt.Printf("[RESV-DEBUG] userId=%d matched %d row(s) - canApprove=%v\n", userId, count, count > 0)
	return count > 0, nil
}

// ApproveReservation signs off on a pending reservation. It's already been holding
// stock since the moment it was created (Pending counts the same as Approved in the
// Reserved sum - see the Status doc comment on StockReservation), so this doesn't touch
// availability at all; it just records who authorized it and moves it off the pending
// queue.
func (s *ItemStockService) ApproveReservation(reservationId uint, approvedByUserId uint) (int, error) {
	return s.setReservationDecision(reservationId, approvedByUserId, inventory_models.ReservationStatusApproved)
}

// RejectReservation declines a pending reservation. Unlike Approve, this DOES free the
// stock back up - Rejected is the one status excluded from the Reserved sum - so it logs
// a RELEASE ledger row the same way a manual release or expiry does.
func (s *ItemStockService) RejectReservation(reservationId uint, rejectedByUserId uint) (int, error) {
	return s.setReservationDecision(reservationId, rejectedByUserId, inventory_models.ReservationStatusRejected)
}

func (s *ItemStockService) setReservationDecision(reservationId uint, actingUserId uint, newStatus string) (int, error) {
	canApprove, err := s.UserCanApproveReservations(actingUserId)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking approver permission")
	}
	if !canApprove {
		return fiber.StatusForbidden, errors.New("this user's position is not authorized to approve or reject reservations")
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

	if reservation.Status != inventory_models.ReservationStatusPending {
		return fiber.StatusConflict, fmt.Errorf("reservation is already %s", reservation.Status)
	}

	now := time.Now()
	reservation.Status = newStatus
	reservation.ApprovedBy = &actingUserId
	reservation.ApprovedAt = &now

	if err := services.DbUpdate(tx, &reservation, map[string]interface{}{"id": reservation.ID}); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed updating reservation")
	}

	if newStatus == inventory_models.ReservationStatusRejected {
		availableBefore, err := s.getAvailableSnapshot(tx, reservation.ItemId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}
		availableAfter := availableBefore + int(reservation.Qty)
		remarks := fmt.Sprintf("Reservation rejected for sales quotation #%d", reservation.QuotationId)
		if err := s.logReservationLedger(tx, reservation.ItemId, "RELEASE", availableBefore, availableAfter, int(reservation.Qty), reservation.SourceType, reservation.SourceId, reservation.QuotationId, fmt.Sprintf("user#%d", actingUserId), remarks); err != nil {
			return fiber.StatusInternalServerError, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return fiber.StatusOK, nil
}

// GetPendingReservations lists every reservation still awaiting a dispatcher/inventory
// manager's decision, oldest first, joined with just enough item/quotation context to
// review without a separate lookup per row.
func (s *ItemStockService) GetPendingReservations() ([]inventory_models.PendingReservationView, int, error) {
	var response []inventory_models.PendingReservationView

	query := `
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
			r.status
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
		) cust
		WHERE r.status = ?
		ORDER BY r.reserved_at ASC
	`

	if err := initializers.DB.Raw(query, inventory_models.ReservationStatusPending).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting pending reservations")
	}

	return response, fiber.StatusOK, nil
}
