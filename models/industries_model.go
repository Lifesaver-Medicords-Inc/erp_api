package models

type IndustriesContent struct {
	Name string `json:"name"`
}

type Industries struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	EntityContent
}

func (Industries) TableName() string {
	return "tbl_setup_bpi_industries"
}

type IndustriesAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	IndustriesContent
	At
}

func (IndustriesAt) TableName() string {
	return "z_tbl_setup_bpi_industries_at"
}
