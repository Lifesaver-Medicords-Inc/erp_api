package accounting_models

import "github.com/pierceperado/smpc/models"

type ChartClassContent struct {
	Type string `json:"type"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
}
type ChartClass struct {
	ID uint `gorm:"primarykey" json:"id"`
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
	Type  string `json:"type"`
	models.At
}

func (ChartClassAt) TableName() string {
	return "z_tbl_setup_chart_class_at"
}
