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
	ItemStocksContent
	models.At
}

func (ItemStocksAt) TableName() string {
	return "z_tbl_inv_item_stocks_at"
}
