package models

type ChildsContent struct {
	ParentId uint   `json:"parent_id"`
	Type     string `json:"type"`
	Model    string `json:"model"`
}

type Childs struct {
	ID uint `gorm:"primarykey" json:"id"`
	ChildsContent
}

func (Childs) TableName() string {
	return "tbl_setup_childss"
}

type ChildsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ChildsContent
	At
}

func (ChildsAt) TableName() string {
	return "z_tbl_setup_childss_at"
}
