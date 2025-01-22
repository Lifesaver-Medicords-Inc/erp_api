package models

type Application struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func (Application) TableName() string {
	return "tbl_setup_application"
}

type ApplicationAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	At    `json:"at"`
}

func (ApplicationAt) TableName() string {
	return "z_tbl_setup_application_at"
}
