package dispatching_models

import "github.com/pierceperado/smpc/models"

type CalendarCostTypeContent struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CalendarCostTypeModel struct {
	ID uint `gorm:"primarykey" json:"id"`
	CalendarCostTypeContent
}

func (CalendarCostTypeModel) TableName() string {
	return "tbl_setup_calendar_cost_type"
}

type CalendarCostTypeAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	CalendarCostTypeContent
	models.At
}

func (CalendarCostTypeAt) TableName() string {
	return "z_tbl_setup_calendar_cost_type_at"
}
