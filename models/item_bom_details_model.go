package models

type SetupItemBomDetailsContent struct {
	ItemBomID uint `json:"item_bom_id"`
	ItemID    uint `json:"item_id"`
	//Size      uint `json:"size"`
	BomQty    uint    `json:"bom_qty"`
	UnitPrice uint    `json:"unit_price"` //should be float32
	NetPrice  float32 `json:"net_price"`
}

type SetupItemBomDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	SetupItemBomDetailsContent
}

func (SetupItemBomDetails) TableName() string {
	return "tbl_setup_item_bom_details"
}

type SetupItemBomDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SetupItemBomDetailsContent
	At
}

func (SetupItemBomDetailsAt) TableName() string {
	return "z_tbl_setup_item_bom_details_at"
}
