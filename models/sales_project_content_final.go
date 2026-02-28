package models

type SalesProjectContentFinalContent struct {
	SalesProjectContentID uint    `json:"sales_project_content_id"`
	Final                 string  `json:"final"`
	Fla                   float64 `json:"fla"`
	Voltage               float64 `json:"voltage"`
}

type SalesProjectContentFinal struct {
	ID uint `json:"id" gorm:"primaryKey"`
	SalesProjectContentFinalContent
}

func (SalesProjectContentFinal) TableName() string {
	return "tbl_trans_sales_project_content_final"
}

type SalesProjectContentFinalAt struct {
	ID    uint `json:"id" gorm:"primaryKey"`
	RefID uint `json:"ref_id"`
	SalesProjectContentFinalContent
	At
}

func (SalesProjectContentFinalAt) TableName() string {
	return "z_tbl_trans_sales_project_content_final_at"
}
