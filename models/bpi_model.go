package models

type BpiContent struct {
	SalesId     uint   `json:"sales_id"`
    Name        string `json:"name"`
    MainWebsite string `json:"main_website"`
    Tin         string `json:"tin"`
    MainTelNo   string `json:"main_tel_no"`
}

type Bpi struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiContent
}

func (Bpi) TableName() string {
	return "tbl_bpi"
}

type BpiAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiContent
	At
}

func (BpiAt) TableName() string {
	return "z_tbl_bpi_at"
}
