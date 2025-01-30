package models

type ApplicationContent struct {
	Name string `json:"name"`
}

type Application struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	ApplicationContent
}

func (Application) TableName() string {
	return "tbl_setup_application"
}

type ApplicationAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	ApplicationContent
	At
}

func (ApplicationAt) TableName() string {
	return "z_tbl_setup_application_at"
}
