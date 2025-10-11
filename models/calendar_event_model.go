package models

import "time"

type CalendarEventContent struct {
	Title       string    `gorm:"size:150" json:"title"`
	Description string    `gorm:"size:255" json:"description"`
	StartTime   time.Time `json:"start"`
	EndTime     time.Time `json:"end"`
	EventType   string    `gorm:"size:50" json:"event_type"`
	CreatedByID uint      `json:"created_by_id"`
}

type CalendarEventModel struct {
	ID             uint  `gorm:"primaryKey" json:"id"`
	DepartmentID   uint  `json:"department_id"`
	RelatedOrderID *uint `json:"related_order_id"`
	CalendarEventContent
}

func (CalendarEventModel) TableName() string {
	return "tbl_calendar_event"
}

type CalendarEventAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	CalendarEventContent
	At
}

func (CalendarEventAt) TableName() string {
	return "z_tbl_calendar_event_at"
}
