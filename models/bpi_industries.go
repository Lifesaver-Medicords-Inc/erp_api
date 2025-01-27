package models

type BpiIndustriesContent struct {
	BpiId      uint `json:"bpi_id"`
	IndustryId uint `json:"industry_id"`
}

type BpiIndustries struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiIndustriesContent
}

func (BpiIndustries) TableName() string {
	return "tbl_bpi_industries"
}

type BpiIndustriesAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiIndustriesContent
	At
}

func (BpiIndustriesAt) TableName() string {
	return "z_tbl_bpi_industries_at"
}
