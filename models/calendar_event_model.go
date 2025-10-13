package models

import "time"

type CalendarEventContent struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
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
