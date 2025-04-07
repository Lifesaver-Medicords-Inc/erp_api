package models

type ItemProductionView struct {
	ItemId    int    `json:"item_id"`
	BomId     int    `json:"bom_id"`
	BomItemId int    `json:"bom_item_id"`
	ItemModel string `json:"item_model"`
	ItemCode  string `json:"item_code"`
	BomQty    int    `json:"bom_qty"`
}

func (ItemProductionView) TableName() string {
	return "vw_item_production_list"
}
