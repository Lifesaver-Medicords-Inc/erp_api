package models

type ClassContent struct {
	Name string `json:"name"`
}
type Class struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	ClassContent
}

func (Class) TableName() string {
	return "tbl_setup_item_class"
}

type ClassAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ClassContent
	At
}

func (ClassAt) TableName() string {
	return "z_tbl_setup_item_class_at"
}
