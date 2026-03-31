package models

type JobOrderContent struct {
	BomId          uint   `json:"bom_id"`
	SoId           uint   `json:"so_id"`
	IrId           uint   `json:"ir_id"`
	OrderDetailsId uint   `json:"order_details_id"`
	Date           string `json:"date"`
	SalesOrder     string `json:"sales_order"`
	GeneralName    string `json:"general_name"`
	ItemDesc       string `json:"item_desc"`
	Type           string `json:"type"`
	Materials      string `json:"materials"`
	Quantity       uint   `json:"quantity"`
	Due            string `json:"due"`
	EngrId         uint   `json:"engr_id"`
	AEngr          string `json:"a_engr"`
	ItemRqst       string `json:"item_rqst"`
	Status         string `json:"status"`
	SerialNo       string `json:"serial_no"`
	ReportBase     string `json:"report_base"`
	Report         string `json:"report"`
}
type JobOrder struct {
	ID uint `gorm:"primarykey" json:"id"`
	JobOrderContent
}

func (JobOrder) TableName() string {
	return "tbl_trans_job_order"
}

type JobOrderAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	JobOrderContent
	At
}

func (JobOrderAt) TableName() string {
	return "z_tbl_trans_job_order_at"
}

type JobOrderSales struct {
	Id           int    `json:"id"`
	Customer     string `json:"customer"`
	Tin          string `json:"tin"`
	Code         string `json:"code"`
	DeliveryTo   string `json:"delivery_to"`
	BillTo       string `json:"bill_to"`
	DocNo        string `json:"doc_no"`
	Date         string `json:"date"`
	DeliveryDate string `json:"delivery_date"`
	ReferenceDoc int    `json:"reference_doc"`
	Status       string `json:"status"`
}

func (JobOrderSales) TableName() string {
	return "vw_get_sales_order_engineering"
}

type JobOrderSalesDetails struct {
	Id       int    `json:"id"`
	SoId     int    `json:"so_id"`
	ItemCode string `json:"item_code"`
	ItemDesc string `json:"item_desc"`
	Stock    int    `json:"stock"`
	ReqQty   int    `json:"req_qty"`
	Remark   string `json:"remark"`
	Status   string `json:"status"`
}

func (JobOrderSalesDetails) TableName() string {
	return "vw_get_sales_order_details_engineering"
}

type Components struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Stock    int    `json:"stock"`
}

type SalesOrderViewBody struct {
	SalesOrderView        []JobOrderSales        `json:"sales_order_view"`
	SalesOrderDetailsView []JobOrderSalesDetails `json:"sales_order_details_view"`
}
