package models

type BomView struct {
	ItemID      uint   `json:"item_id"`
	ShortDesc   string `json:"short_desc"`
	ItemCode    string `json:"item_code"`
	GeneralName string `json:"general_name"`
	ItemModel   string `json:"item_model"`
	UomName     string `json:"uom_name"`
}

func (BomView) TableName() string {
	return "vw_get_item_bom"
}
