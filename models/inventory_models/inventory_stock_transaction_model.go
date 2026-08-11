package inventory_models

import "time"

// StockTransaction is one immutable row per stock movement on tbl_inv_item_stocks.
// Unlike z_tbl_inv_item_stocks_at (written by item_stock_service.go on the paths that
// go through it), every row here is written by the tr_inv_item_stocks_ledger trigger
// (see sql/triggers/tr_inv_item_stocks_ledger.sql) directly on INSERT/UPDATE/DELETE
// against tbl_inv_item_stocks - so it's guaranteed to exist even for a write that
// bypasses the shared service functions, which is exactly what happens today in
// receiving_report_service.go's and pick_activity_service2.go's delete-reversal blocks.
//
// qty_before/qty_after/qty_change/direction are always populated by the trigger from
// the raw column values, with no dependency on application code. source_type/
// source_id/remarks are enrichment only: they're read from SESSION_CONTEXT, which the
// app sets right before a write via services.SetStockAuditContext (see
// item_stock_service.go). A write that doesn't set that context still produces a
// ledger row - just with those three columns NULL - so the running balance is never
// silently lost, only the "why" is.
type StockTransaction struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	RefId       uint   `gorm:"column:ref_id" json:"ref_id"` // tbl_inv_item_stocks.id
	ItemId      uint   `json:"item_id"`
	WarehouseId uint   `json:"warehouse_id"`
	BinLocation string `json:"bin_location"`
	StockUom    string `json:"stock_uom"`
	DocNo       int    `json:"doc_no"`

	Direction string `json:"direction"` // IN | OUT
	QtyBefore int    `json:"qty_before"`
	QtyAfter  int    `json:"qty_after"`
	QtyChange int    `json:"qty_change"` // signed: qty_after - qty_before

	// Enrichment - populated from SESSION_CONTEXT when the app sets it, NULL otherwise.
	SourceType *string `json:"source_type"`
	SourceId   *uint   `json:"source_id"`
	Remarks    *string `json:"remarks"`

	// Cost enrichment - only populated on movements that went through FIFO lot tracking
	// (see ConsumeLotsFIFO/CreateStockLot in item_stock_services). NULL for movements
	// with no lot history (manual add/adjust, or stock that predates this feature).
	UnitCost     *float64 `json:"unit_cost"`
	SupplierId   *uint    `json:"supplier_id"`
	Supplier     *string  `json:"supplier"`
	PurchaseDate *string  `json:"purchase_date"`

	// Always populated by the trigger itself, independent of the app.
	TransactionAt time.Time `gorm:"column:transaction_at" json:"transaction_at"`
	DbUser        string    `gorm:"column:db_user" json:"db_user"`
}

func (StockTransaction) TableName() string {
	return "tbl_inv_stock_transactions"
}

// StockTransactionListView is the display shape for a stock ledger screen - one row
// per movement, joined with human-readable item/warehouse names instead of raw IDs
// (same pattern as ItemStockListView in inventory_item_stock_model.go).
type StockTransactionListView struct {
	ID            uint    `json:"id"`
	ItemId        uint    `json:"item_id"`
	ItemCode      string  `json:"item_code"`
	ItemName      string  `json:"item_name"`
	WarehouseId   uint    `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	BinLocation   string  `json:"bin_location"`
	Direction     string  `json:"direction"`
	QtyBefore     int     `json:"qty_before"`
	QtyAfter      int     `json:"qty_after"`
	QtyChange     int     `json:"qty_change"`
	DocNo         int     `json:"doc_no"`
	SourceType    *string   `json:"source_type"`
	SourceId      *uint     `json:"source_id"`
	Remarks       *string   `json:"remarks"`
	UnitCost      *float64  `json:"unit_cost"`
	SupplierId    *uint     `json:"supplier_id"`
	Supplier      *string   `json:"supplier"`
	PurchaseDate  *string   `json:"purchase_date"`
	TransactionAt time.Time `json:"transaction_at"`
	DbUser        string    `json:"db_user"`
}
