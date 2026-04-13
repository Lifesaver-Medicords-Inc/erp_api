package inventory_models

type ItemStockAndLocationView struct {
	ItemId      uint   `json:"item_id"`
	BinLocation string `json:"bin_location"`
	StockQty    int    `json:"stock_qty"`
	StockUom    string `json:"stock_uom"`
	WarehouseId uint   `json:"warehouse_id"`
}
