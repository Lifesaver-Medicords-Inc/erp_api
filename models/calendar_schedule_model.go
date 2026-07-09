package models

type CalendarScheduleContent struct {
	DepartmentType string `json:"department_type"` // Sales, Engineering, Logistics
	Title          string `json:"title"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	CategoryId     uint   `json:"category_id"`
	Description    string `json:"description"`
	Location       string `json:"location"`
	People         string `json:"people"`
	VehicleId      uint   `json:"vehicle_id"`
	Notes          string `json:"notes"`
}

type CalendarScheduleModel struct {
	ID             uint  `gorm:"primaryKey" json:"id"`
	DepartmentID   uint  `json:"department_id"`
	ReferenceDocId *uint `json:"reference_doc_id,omitempty"`
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
