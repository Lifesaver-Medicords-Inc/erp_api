package models

type SetupItemBomContent struct {
	ItemID         uint   `json:"item_id"`
	ProductionQty  uint   `json:"production_qty"`
	ProductionType string `json:"production_type"`
	Labor          string `json:"labor"`
}

type SetupItemBom struct {
	ID uint `gorm:"primarykey" json:"id"`
	SetupItemBomContent
}

func (SetupItemBom) TableName() string {
	return "tbl_setup_item_bom"
}

type SetupItemBomAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SetupItemBomContent
	At
}

func (SetupItemBomAt) TableName() string {
	return "z_tbl_setup_item_bom_at"
}
