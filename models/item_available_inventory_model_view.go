package models

type ItemAvailableInventoryModelView struct {
	WarehouseId   int     `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Location      string  `json:"location"`
	ItemID        uint    `json:"item_id"`
	StockQty      float32 `json:"stock_qty"`
	StockUom      string  `json:"stock_uom"`
}

func (ItemAvailableInventoryModelView) TableName() string {
	return "vw_get_item_available_inventory"
}
