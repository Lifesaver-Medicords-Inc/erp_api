package models

type BomView struct {
	ItemID      uint   `json:"item_id"`
	ShortDesc   string `json:"short_desc"`
	ItemCode    string `json:"item_code"`
	GeneralName string `json:"general_name"`
	ItemModel   string `json:"item_model"`
	UomName     string `json:"uom_name"`
	Size        string `json:"size"`
}

func (BomView) TableName() string {
	return "vw_get_item_bom"
}

type AllBomView struct {
	ItemID      uint   `json:"item_id"`
	ShortDesc   string `json:"short_desc"`
	ItemCode    string `json:"item_code"`
	GeneralName string `json:"general_name"`
	ItemModel   string `json:"item_model"`
	UomName     string `json:"uom_name"`
	Size        string `json:"size"`
}

func (AllBomView) TableName() string {
	return "vw_get_all_item_bom"
}
