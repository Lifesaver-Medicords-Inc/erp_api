package models

type ParentContent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Parent struct {
	ID uint `gorm:"primarykey" json:"id"`
	ParentContent
}

func (Parent) TableName() string {
	return "tbl_setup_parents"
}

type ParentAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ParentContent
	At
}

func (ParentAt) TableName() string {
	return "z_tbl_setup_parents_at"
}
