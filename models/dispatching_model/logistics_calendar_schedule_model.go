package dispatching_models

type LogisticsCalendarScheduleContent struct {
	// true = Internal (own vehicle/driver, multiple routes with cost tracking)
	// false = External (third-party courier, single simple record)
	IsExternal           bool    `json:"is_external"`
	SalesOrderDocNo      string  `json:"reference_doc_no"`
	SalesOrderId         uint    `json:"sales_order_id"`
	DeliveryReceiptDocNo string  `json:"delivery_receipt_doc_no"`
	DeliveryReceiptId    uint    `json:"delivery_receipt_id"`
	SalesInvoiceDocNo    string  `json:"sales_invoice_doc_no"`
	Category             string  `json:"category"`
	ClientSupplier       string  `json:"client_supplier"`
	DriverName           string  `json:"driver_name"`
	VehicleId            uint    `json:"vehicle_id"`
	Trucking             string  `json:"trucking"`
	Courier              string  `json:"courier"`
	PickupTime           string  `json:"pickup_time"`
	ArrivalTime          string  `json:"arrival_time"`
}

type LogisticsCalendarScheduleModel struct {
	CalendarScheduleBase
	LogisticsCalendarScheduleContent
	Routes []LogisticsRoute `gorm:"foreignKey:ScheduleId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"routes,omitempty"`
}

func (LogisticsCalendarScheduleModel) TableName() string {
	return "tbl_dispatching_logistics_calendar_schedule"
}

type LogisticsCalendarScheduleModelAt struct {
	CalendarSchedulesBaseAt
	LogisticsCalendarScheduleContent
}

func (LogisticsCalendarScheduleModelAt) TableName() string {
	return "z_tbl_dispatching_logistics_calendar_schedule_at"
}
