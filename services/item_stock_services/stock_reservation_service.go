package item_stock_services

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// CreateStockReservation places a soft hold for a quotation line. expiresAt is
// whatever the caller resolved from the parent document (e.g. a quotation's
// ValidUntil) - pass nil if it couldn't be parsed, but note that means
// ExpireStockReservations will never clean this row up on its own.
func (s *ItemStockService) CreateStockReservation(tx *gorm.DB, itemId uint, qty uint, sourceType string, sourceId uint, quotationId uint, expiresAt *time.Time) error {
	if qty == 0 {
		return nil
	}

	reservation := &inventory_models.StockReservation{
		ItemId:      itemId,
		Qty:         qty,
		SourceType:  sourceType,
		SourceId:    sourceId,
		QuotationId: quotationId,
		ReservedAt:  time.Now(),
		ExpiresAt:   expiresAt,
	}

	return services.DbInsert(tx, reservation)
}

// ReleaseStockReservation removes the hold for one source line (e.g. a quotation line
// being deleted). A row simply not existing means "not reserved" - there's no
// is_active flag to flip.
func (s *ItemStockService) ReleaseStockReservation(tx *gorm.DB, sourceType string, sourceId uint) error {
	return services.DbDelete(tx, &inventory_models.StockReservation{}, map[string]interface{}{
		"source_type": sourceType,
		"source_id":   sourceId,
	})
}

// ExpireStockReservations deletes every reservation whose ExpiresAt has passed.
// Nothing in this codebase calls this on its own - see initializers.StartReservationSweep
// for the periodic goroutine that does.
func (s *ItemStockService) ExpireStockReservations(tx *gorm.DB) (int64, error) {
	result := tx.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&inventory_models.StockReservation{})
	return result.RowsAffected, result.Error
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
			WHERE (? = 0 OR item_id = ?)
			GROUP BY item_id
		) r ON r.item_id = p.item_id
	`

	if err := initializers.DB.Raw(query, itemId, itemId, itemId, itemId).Scan(&response).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting available stock")
	}

	return response, fiber.StatusOK, nil
}
