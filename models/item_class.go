package models

type Class struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;not null" json:"code"`
	Name string `json:"name"`
}

func (Class) TableName() string {
	return "tbl_setup_item_class"
}

type ClassAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	At    `json:"at"`
}

func (ClassAt) TableName() string {
	return "z_tbl_setup_item_class_at"
}
