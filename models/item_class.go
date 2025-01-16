package models

type Class struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func (Class) TableName() string {
	return "tbl_setup_item_class"
}

type ClassAt struct {
	Class
	At
}

func (ClassAt) TableName() string {
	return "z_tbl_setup_item_class_at"
}
