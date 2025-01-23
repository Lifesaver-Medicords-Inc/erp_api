package models

type TypeContent struct {
	Name string `json:"name"`
}
type Type struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	TypeContent
}

func (Type) TableName() string {
	return "tbl_setup_item_type"
}

type TypeAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	TypeContent
	At
}

func (TypeAt) TableName() string {
	return "z_tbl_setup_item_type_at"
}
