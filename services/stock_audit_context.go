package services

import (
	"github.com/pierceperado/smpc/models/inventory_models"
	"gorm.io/gorm"
)

// SetStockAuditContext tells the tr_inv_item_stocks_ledger trigger (see
// sql/triggers/tr_inv_item_stocks_ledger.sql) why the write about to happen on
// tbl_inv_item_stocks is occurring, via SQL Server SESSION_CONTEXT. Call this once on
// the active transaction, right before whichever DbInsert/DbUpdate call touches
// tbl_inv_item_stocks - the trigger reads it back inside the same trigger invocation.
//
// This is enrichment only: the trigger writes a ledger row (qty_before/qty_after/
// qty_change/direction) regardless of whether this was called. Skipping it just means
// that row's source_type/source_id/remarks/cost columns come back NULL.
//
// cost is optional (pass nil when there's no FIFO lot data for this movement, e.g.
// manual add/adjust) - see ConsumeLotsFIFO/CreateStockLot in item_stock_services for
// who populates it.
func SetStockAuditContext(tx *gorm.DB, sourceType string, sourceId uint, remarks string, cost *inventory_models.LotInfo) error {
	if err := tx.Exec(
		"EXEC sp_set_session_context @key = N'stock_source_type', @value = ?",
		sourceType,
	).Error; err != nil {
		return err
	}

	if err := tx.Exec(
		"EXEC sp_set_session_context @key = N'stock_source_id', @value = ?",
		sourceId,
	).Error; err != nil {
		return err
	}

	if remarks != "" {
		if err := tx.Exec(
			"EXEC sp_set_session_context @key = N'stock_remarks', @value = ?",
			remarks,
		).Error; err != nil {
			return err
		}
	}

	if cost == nil {
		return nil
	}

	if err := tx.Exec(
		"EXEC sp_set_session_context @key = N'stock_unit_cost', @value = ?",
		cost.UnitCost,
	).Error; err != nil {
		return err
	}

	if err := tx.Exec(
		"EXEC sp_set_session_context @key = N'stock_supplier_id', @value = ?",
		cost.SupplierId,
	).Error; err != nil {
		return err
	}

	if err := tx.Exec(
		"EXEC sp_set_session_context @key = N'stock_supplier', @value = ?",
		cost.Supplier,
	).Error; err != nil {
		return err
	}

	return tx.Exec(
		"EXEC sp_set_session_context @key = N'stock_purchase_date', @value = ?",
		cost.PurchaseDate,
	).Error
}
