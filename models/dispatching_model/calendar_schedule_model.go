package dispatching_models

import "github.com/pierceperado/smpc/models"

type CalendarScheduleBase struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Department  string `json:"department"`
	CategoryId  uint   `json:"category_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	People      string `json:"people"`
	Notes       string `json:"notes"`
}

type CalendarSchedulesBaseAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	models.At
}
