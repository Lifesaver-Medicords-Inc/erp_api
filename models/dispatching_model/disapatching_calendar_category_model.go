package dispatching_models

import "github.com/pierceperado/smpc/models"

type CalendarCategoryContent struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CalendarCategoryModel struct {
	ID uint `gorm:"primarykey" json:"id"`
	CalendarCategoryContent
}

func (CalendarCategoryModel) TableName() string {
	return "tbl_setup_calendar_category"
}

type CalendarCategoryAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	CalendarCategoryContent
	models.At
}

func (CalendarCategoryAt) TableName() string {
	return "z_tbl_setup_calendar_category_at"
}
