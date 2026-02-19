package dispatching_models

type LogisticsCalendarScheduleContent struct {
	SalesOrderDocNo      string `json:"reference_doc_no"`
	SalesOrderId         uint   `json:"sales_order_id"`
	DeliveryReceiptDocNo string `json:"delivery_receipt_doc_no"`
	DeliveryReceiptId    uint   `json:"delivery_receipt_id"`
	DriverName           string `json:"driver_name"`
	VehicleId            uint   `json:"vehicle_id"`
	Trucking             string `json:"trucking"`
	Courier              string `json:"courier"`
}

type LogisticsCalendarScheduleModel struct {
	CalendarScheduleBase
	SalesCalendarScheduleContent
}

func (LogisticsCalendarScheduleModel) TableName() string {
	return "tbl_dispatching_logistics_calendar_schedule"
}

type LogisticsCalendarScheduleModelAt struct {
	CalendarSchedulesBaseAt
	SalesCalendarScheduleContent
}

func (LogisticsCalendarScheduleModelAt) TableName() string {
	return "z_tbl_dispatching_logistics_calendar_schedule_at"
}
