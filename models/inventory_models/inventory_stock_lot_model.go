package inventory_models

import "time"

// StockLot is one purchase "batch" for an item+warehouse+bin - created whenever
// UpsertStockWithTx runs with lot info attached (currently only Receiving Report does
// this; see item_stock_service.go). QtyRemaining starts equal to QtyReceived and is
// drawn down by ConsumeLotsFIFO whenever a sale (DeductStockWithTx) needs stock from
// this item+bin, oldest lot first. tbl_inv_item_stocks' pooled balance for a given
// item+bin should always equal the sum of QtyRemaining across its lots - that's the
// sanity check this table is meant to satisfy, not replace.
type StockLot struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ItemId      uint      `json:"item_id"`
	WarehouseId uint      `json:"warehouse_id"`
	BinLocation string    `json:"bin_location"`
	UnitCost    float64   `json:"unit_cost"`
	SupplierId  uint      `json:"supplier_id"`
	Supplier    string    `json:"supplier"`
	PurchaseDate string   `json:"purchase_date"`
	QtyReceived  int      `json:"qty_received"`
	QtyRemaining int      `json:"qty_remaining"`
	// Where this lot came from - almost always "receiving_report" today, but kept
	// generic (rather than a hard FK) since manual stock adds could create a
	// zero-cost lot too (see AdjustItemStock/InsertItemStock - not wired up yet).
	SourceType string    `json:"source_type"`
	SourceId   uint       `json:"source_id"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (StockLot) TableName() string {
	return "tbl_inv_stock_lots"
}

// StockLotConsumption records exactly how much of which lot a single deduction event
// drew from. RefType/RefId reuse the same source_type/source_id passed to
// DeductStockWithTx (the parent document - e.g. an Item Request or Pick Activity's own
// ID), not a per-line identifier. That's precise enough for reversal in the normal
// case, with one known edge case: a document with two lines for the exact same
// item+bin would have both lines' consumptions attributed to the same ref, so
// releasing one would release both. Not expected to come up often; flagged here rather
// than solved, to avoid threading a finer-grained ref through every call site.
//
// This table only exists so a reversal (RestoreStockWithTx) can put qty back onto the
// exact lot(s) it was taken from, in the exact amounts - see ReleaseLotsFIFO.
type StockLotConsumption struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	LotId       uint      `json:"lot_id"`
	RefType     string    `json:"ref_type"`
	RefId       uint      `json:"ref_id"`
	QtyConsumed int       `json:"qty_consumed"`
	CreatedAt   time.Time `json:"created_at"`
}

func (StockLotConsumption) TableName() string {
	return "tbl_inv_stock_lot_consumptions"
}

// LotInfo carries purchase details into UpsertStockWithTx so it can create a StockLot
// alongside the usual tbl_inv_item_stocks balance update. Pass nil for any caller that
// isn't a real purchase (manual stock add has no PO, so no cost to record yet).
type LotInfo struct {
	UnitCost     float64
	SupplierId   uint
	Supplier     string
	PurchaseDate string
	SourceType   string
	SourceId     uint
}
