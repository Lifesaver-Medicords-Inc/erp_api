package models

type BomViewList struct {
	ID             uint   `json:"id"`
	ItemID         uint   `json:"item_id"`
	ProductionQty  uint   `json:"production_qty"`
	ProductionType string `json:"production_type"`
	Labor          string `json:"labor"`
	ItemModel      string `json:"item_model"`
	ItemCode       string `json:"item_code"`
	GeneralName    string `json:"general_name"`
	ProductionCost uint   `json:"production_cost"`
}

func (BomViewList) TableName() string {
	return "vw_get_bom_list"
}
