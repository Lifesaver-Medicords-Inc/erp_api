package models

type ChartClassContent struct {
	Type string `json:"type"`
}
type ChartClass struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
	ChartClassContent
}

func (ChartClass) TableName() string {
	return "tbl_setup_chart_class"
}

type ChartClassAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	ChartClassContent
	At
}

func (ChartClassAt) TableName() string {
	return "z_tbl_setup_chart_class_at"
}
