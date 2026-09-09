package services

import (
	"strconv"

	"github.com/pierceperado/smpc/models/inventory_models"
	"gorm.io/gorm"
)

// SetStockAuditContext tells the tr_inv_item_stocks_ledger trigger (see
// sql/triggers/tr_inv_item_stocks_ledger.sql) why the write about to happen on
// tbl_inv_item_stocks is occurring. Call this once on the active transaction,
// right before whichever DbInsert/DbUpdate call touches tbl_inv_item_stocks -
// the trigger reads it back inside the same trigger invocation.
//
// This is enrichment only: the trigger writes a ledger row (qty_before/qty_after/
// qty_change/direction) regardless of whether this was called. Skipping it just means
// that row's source_type/source_id/remarks/cost columns come back NULL.
//
// cost is optional (pass nil when there's no FIFO lot data for this movement, e.g.
// manual add/adjust) - see ConsumeLotsFIFO/CreateStockLot in item_stock_services for
// who populates it.
//
// Carried through tbl_inv_stock_audit_context, keyed by @@SPID, rather than
// SQL Server's SESSION_CONTEXT - see StockAuditContext for why (SESSION_CONTEXT
// is 2016+, the production server is 2012).
//
// Every column is written on every call, including the ones this caller has
// nothing for. The previous SESSION_CONTEXT version skipped the remarks key when
// remarks was empty and all four cost keys when cost was nil, which left the
// values from an earlier movement in place - and since the context is scoped to
// the session, not the statement, the next write on that pooled connection was
// attributed to whatever came before it. Writing the full row makes each call
// replace the last completely.
func SetStockAuditContext(tx *gorm.DB, sourceType string, sourceId uint, remarks string, cost *inventory_models.LotInfo) error {
	ctx := inventory_models.StockAuditContext{
		SourceType: sourceType,
		SourceId:   strconv.FormatUint(uint64(sourceId), 10),
		Remarks:    remarks,
	}

	if cost != nil {
		ctx.UnitCost = strconv.FormatFloat(cost.UnitCost, 'f', -1, 64)
		ctx.SupplierId = strconv.FormatUint(uint64(cost.SupplierId), 10)
		ctx.Supplier = cost.Supplier
		ctx.PurchaseDate = cost.PurchaseDate
	}

	// DELETE + INSERT rather than an upsert: MERGE on 2012 needs care around
	// concurrent sessions, and this pair only ever touches this session's own
	// row, so two sessions cannot contend. Both statements run inside the
	// caller's transaction, on the caller's connection - which is what makes
	// @@SPID the correct key.
	if err := tx.Exec(
		"DELETE FROM tbl_inv_stock_audit_context WHERE spid = @@SPID",
	).Error; err != nil {
		return err
	}

	return tx.Exec(
		`INSERT INTO tbl_inv_stock_audit_context
		     (spid, source_type, source_id, remarks, unit_cost, supplier_id, supplier, purchase_date)
		 VALUES (@@SPID, ?, ?, ?, ?, ?, ?, ?)`,
		ctx.SourceType, ctx.SourceId, ctx.Remarks,
		ctx.UnitCost, ctx.SupplierId, ctx.Supplier, ctx.PurchaseDate,
	).Error
}
