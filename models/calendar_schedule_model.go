package models

import "time"

type CalendarScheduleContent struct {
	DepartmentType string    `json:"department_type"` // Sales, Engineering, Logistics
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	ReferenceType  string    `json:"reference_type"` // e.g., "SalesOrder", "DeliveryReceipt"
	ReferenceId    uint      `json:"reference_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CalendarScheduleModel struct {
	ID             uint  `gorm:"primaryKey" json:"id"`
	DepartmentID   uint  `json:"department_id"`
	RelatedOrderID *uint `json:"related_order_id"`
	CalendarScheduleContent
}

func (CalendarScheduleModel) TableName() string {
	return "tbl_calendar_schedule"
}

type CalendarScheduleAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	CalendarScheduleContent
	At
}

func (CalendarScheduleAt) TableName() string {
	return "z_tbl_calendar_schedule_at"
}
