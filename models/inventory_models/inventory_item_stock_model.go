package inventory_models

import "github.com/pierceperado/smpc/models"

type ItemStocksContent struct {
	ItemId      uint   `json:"item_id"`
	StockQty    *int   `json:"stock_qty"`
	StockUom    string `json:"stock_uom"`
	WarehouseId uint   `json:"warehouse_id"`
	BinLocation string `json:"bin_location"`
	IsActive    *bool  `json:"is_active"`
}

type ItemStocks struct {
	ID    uint `gorm:"primarykey" json:"id"`
	DocNo int  `json:"doc_no"`
	ItemStocksContent
}

func (ItemStocks) TableName() string {
	return "tbl_inv_item_stocks"
}

type ItemStocksAt struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	RefId      uint   `json:"ref_id"`
	Code       string `json:"code"`
	SourceId   uint   `json:"source_id"`
	SourceType string `json:"source_type"`
	// Only meaningful on manual adjustments made from the Inventory Item Stocks module
	// (see AdjustItemStock) - lets the audit trail record WHY a balance was corrected by
	// hand (e.g. "physical count", "damaged goods"), not just what it changed to. Left
	// blank for every other source (receiving, pick activity, etc.).
	Remarks string `json:"remarks"`
	ItemStocksContent
	models.At
}

func (ItemStocksAt) TableName() string {
	return "z_tbl_inv_item_stocks_at"
}

// ItemStockListView is the display shape for the Inventory Item Stocks module's list
// screen (and any other caller, e.g. Sales Order's stock check) - one row per
// item+warehouse+bin, joined with human-readable item/warehouse names instead of raw IDs.
type ItemStockListView struct {
	ID            uint   `json:"id"`
	ItemId        uint   `json:"item_id"`
	ItemCode      string `json:"item_code"`
	ItemName      string `json:"item_name"`
	ItemModel     string `json:"item_model"`
	Brand         string `json:"brand"`
	WarehouseId   uint   `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
	BinLocation   string `json:"bin_location"`
	StockQty      int    `json:"stock_qty"`
	StockUom      string `json:"stock_uom"`
	IsActive      bool   `json:"is_active"`
}

// ItemStockAdjustmentBody is the request body for a manual stock correction. Unlike the
// receive/deduct flows elsewhere in this package (which add/subtract a delta as part of a
// transaction), this SETS stock_qty directly to whatever the user physically counted.
type ItemStockAdjustmentBody struct {
	ID      uint   `json:"id"`
	NewQty  int    `json:"new_qty"`
	Remarks string `json:"remarks"`
}

// StockTransferBody is the request body for §10.6's "Transfer" function - move some or
// all of one bin's stock to a different bin, warehouse-to-warehouse moves included.
// Deliberately no reference document field anywhere on this struct - §10.6 is explicit
// that Stock Transfer is the one stock movement with no document behind it.
type StockTransferBody struct {
	SourceStockId   uint   `json:"source_stock_id"` // tbl_inv_item_stocks.id to transfer FROM
	TransferQty     int    `json:"transfer_qty"`
	DestWarehouseId uint   `json:"dest_warehouse_id"`
	DestBinLocation string `json:"dest_bin_location"`
	Remarks         string `json:"remarks"`
}
