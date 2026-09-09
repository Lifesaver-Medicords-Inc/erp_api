package inventory_models

// StockAuditContext carries the "why" of a pending write to tbl_inv_item_stocks
// across to the tr_inv_item_stocks_ledger trigger, which cannot see the Go call
// stack and has only the inserted/deleted pseudo-tables to work from.
//
// This replaces SQL Server's SESSION_CONTEXT, which the trigger used until
// 2026-09-08. SESSION_CONTEXT and sp_set_session_context are SQL Server 2016 and
// later; the production server is 2012, where the trigger failed to compile at
// all ("Must declare the scalar variable @srcType" - the DECLARE lines call an
// unrecognised function, so every later reference cascades). Note that a
// database's compatibility level does NOT gate that syntax: on a 2016+ engine it
// runs happily at level 110, which is why this survived testing on the dev box.
//
// CONTEXT_INFO() is the usual 2012 substitute but caps at 128 bytes, and these
// values run to roughly a thousand characters - hence a real table.
//
// Keyed by @@SPID, the session id, which is what makes this work: a write and
// the trigger it fires are always the same session, so the trigger reads back
// exactly what the caller wrote. Rows are overwritten in place per session and
// bounded by the connection pool, so the table stays tiny.
type StockAuditContext struct {
	// The session that wrote this row. One row per session, replaced on each
	// call rather than appended.
	Spid int `gorm:"primaryKey;autoIncrement:false" json:"spid"`

	// Stored as text and cast by the trigger (TRY_CAST to INT / DECIMAL), which
	// keeps the previous SESSION_CONTEXT behaviour: a value that will not
	// convert lands as NULL rather than failing the stock write.
	SourceType   string `gorm:"size:100" json:"source_type"`
	SourceId     string `gorm:"size:50" json:"source_id"`
	Remarks      string `gorm:"size:500" json:"remarks"`
	UnitCost     string `gorm:"size:50" json:"unit_cost"`
	SupplierId   string `gorm:"size:50" json:"supplier_id"`
	Supplier     string `gorm:"size:200" json:"supplier"`
	PurchaseDate string `gorm:"size:50" json:"purchase_date"`
}

func (StockAuditContext) TableName() string {
	return "tbl_inv_stock_audit_context"
}
