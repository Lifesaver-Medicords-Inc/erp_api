package models

type ItemBoqDetailsContent struct {
	QQID    uint   `json:"qq_id"`
	ItemsID uint   `json:"items_id"`
	Remarks string `json:"remarks"`
	Notes   string `json:"notes"`
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

type QQView struct {
	QQID          uint   `json:"qq_id"`
	BasedID       uint   `json:"based_id"`
	Components    string `json:"components"`
	Model         string `json:"model"`
	ItemID        uint   `json:"item_id"`
	BomID         uint   `json:"bom_id"`
	IsChild       bool   `json:"is_child"`
	Qty           uint   `json:"qty"`
	UnitOfMeasure string `json:"unit_of_measure"`
	Remarks       string `json:"remarks"`
	Notes         string `json:"notes"`
	CustomerName  string `json:"customer_name"`
	QQNoteID      uint   `json:"qq_note_id"`
}

// test
// test
func (QQView) TableName() string {
	return "vw_qqnotes"
}
