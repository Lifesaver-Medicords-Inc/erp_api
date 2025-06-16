package models

type ChartGroup struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
}

func (ChartGroup) TableName() string {
	return "tbl_setup_chart_group"
}

type ChartGroupAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	At
}

func (ChartGroupAt) TableName() string {
	return "z_tbl_setup_chart_group_at"
}
