package models

type NameContent struct {
	Name string `json:"name"`
}
type Name struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;not null" json:"code"`
	ClassContent
}

func (Name) TableName() string {
	return "tbl_setup_item_name"
}

type NameAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	NameContent
	At
}

func (NameAt) TableName() string {
	return "z_tbl_setup_item_name_at"
}
