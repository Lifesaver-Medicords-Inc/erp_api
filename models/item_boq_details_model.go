package models

type ItemBoqDetailsContent struct {
	ItemBoqID uint   `json:"item_boq_id"`
	ItemsID   uint   `json:"items_id"`
	Remarks   string `json:"remarks"`
	Notes     string `json:"notes"`
}

type ItemBoqDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemBoqDetailsContent
}

func (ItemBoqDetails) TableName() string {
	return "tbl_setup_item_boq_details"
}

type ItemBoqDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemBoqDetailsContent
	At
}

func (ItemBoqDetailsAt) TableName() string {
	return "z_tbl_setup_item_boq_details_at"
}
