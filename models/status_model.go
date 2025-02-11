package models

type StatusContent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Status struct {
	ID uint `gorm:"primarykey" json:"id"`
	StatusContent
}

func (Status) TableName() string {
	return "tbl_setup_status"
}

type StatusAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	StatusContent
	At
}

func (StatusAt) TableName() string {
	return "z_tbl_setup_status_at"
}
