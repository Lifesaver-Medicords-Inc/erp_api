package models

type AllBinLocationView struct {
	Location    string `json:"location"`
	WarehouseId uint   `json:"warehouse_id"`
	StockQty    uint   `json:"stock_qty"`
	StockUom    string `json:"stock_uom"`
}

func (AllBinLocationView) TableName() string {
	return "vw_get_all_bin_location"
}
