package dispatching_models

type EngineeringCalendarScheduleContent struct {
	Driver    string `json:"driver"`
	VehicleId uint   `json:"vehicle_id"`
}

type EngineeringCalendarScheduleModel struct {
	CalendarScheduleBase
	EngineeringCalendarScheduleContent
}

func (EngineeringCalendarScheduleModel) TableName() string {
	return "tbl_dispatching_engineering_calendar_schedule"
}

type EngineeringCalendarScheduleModelAt struct {
	CalendarSchedulesBaseAt
	EngineeringCalendarScheduleContent
}

func (EngineeringCalendarScheduleModelAt) TableName() string {
	return "z_tbl_dispatching_engineering_calendar_schedule_at"
}
