package models

type Name struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;not null" json:"code"`
	Name string `json:"name"`
}

func (Name) TableName() string {
	return "tbl_setup_item_name"
}

type NameAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	At    `json:"at"`
}

func (NameAt) TableName() string {
	return "z_tbl_setup_item_name_at"
}
