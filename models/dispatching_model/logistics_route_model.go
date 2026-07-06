package dispatching_models

import "github.com/pierceperado/smpc/models"

// One leg of an Internal logistics schedule (e.g. a delivery then a pickup).
// A schedule can have multiple routes, each with its own cost breakdown.
type LogisticsRouteContent struct {
	ScheduleId         uint   `gorm:"not null;index" json:"schedule_id"`
	SortOrder          int    `json:"sort_order"`
	ShipType           string `json:"ship_type"`
	ReferenceDoc       string `json:"reference_doc"`
	DeliveryReceiptDoc string `json:"delivery_receipt_doc"`
	SalesInvoiceDoc    string `json:"sales_invoice_doc"`
	ClientSupplier     string `json:"client_supplier"`
	Location           string `json:"location"`
	Receiver           string `json:"receiver"`
	ContactNo          string `json:"contact_no"`
	DepartedAt         string `json:"departed_at"`
	ArrivedAt          string `json:"arrived_at"`
	ReturnedAt         string `json:"returned_at"`
	Notes              string `json:"notes"`
}

type LogisticsRoute struct {
	ID uint `gorm:"primaryKey" json:"id"`
	LogisticsRouteContent
	Costs []LogisticsRouteCost `gorm:"foreignKey:RouteId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"costs,omitempty"`
}

func (LogisticsRoute) TableName() string {
	return "tbl_dispatching_logistics_route"
}

type LogisticsRouteAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	LogisticsRouteContent
	models.At
}

func (LogisticsRouteAt) TableName() string {
	return "z_tbl_dispatching_logistics_route_at"
}
