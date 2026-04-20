package inventory_models

type ItemLocationView struct {
	BinId       int    `json:"bin_id"`
	BinLocation string `json:"bin_location"`
	WarehouseId int    `json:"warehouse_id"`
	ItemId      int    `json:"item_id"`
	StockQty    int    `json:"stock_qty"`
	StockUom    string `json:"stock_uom"`
}
