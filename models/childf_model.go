package models

type ChildfContent struct {
	ParentId    uint   `json:"parent_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Childf struct {
	ID uint `gorm:"primarykey" json:"id"`
	ChildfContent
}

func (Childf) TableName() string {
	return "tbl_setup_childfs"
}

type ChildfAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ChildfContent
	At
}

func (ChildfAt) TableName() string {
	return "z_tbl_setup_childfs_at"
}
