package dispatching_models

type SalesCalendarScheduleContent struct {
	ReferenceDocNo string `json:"reference_doc_no"`
	ReferenceId    uint   `json:"reference_id"`
}

type SalesCalendarScheduleModel struct {
	CalendarScheduleBase
	SalesCalendarScheduleContent
}

func (SalesCalendarScheduleModel) TableName() string {
	return "tbl_dispatching_sales_calendar_schedule"
}

type SalesCalendarScheduleModelAt struct {
	CalendarSchedulesBaseAt
	SalesCalendarScheduleContent
}

func (SalesCalendarScheduleModelAt) TableName() string {
	return "z_tbl_dispatching_sales_calendar_schedule_at"
}
