package models

type JobOrderContent struct {
	Type     string `json:"type"`
	ItemDesc string `json:"item_desc"`
}

type JobOrder struct {
	JobOrderID     uint   `gorm:"primaryKey" json:"job_order_id"`
	BomId          uint   `json:"bom_id"`
	OrderDetailsId uint   `json:"order_details_id"`
	Date           string `json:"date"`
	SalesOrder     string `json:"sales_order"`
	Materials      string `json:"materials"`
	Quantity       uint   `json:"quantity"`
	Due            string `json:"due"`
	A_Engr         string `json:"a_engr"`
	ItemRqst       string `json:"item_rqst"`
	Status         string `json:"status"`
	GeneralName    string `json:"general_name"`
	SerialNo       string `json:"serial_no"`
	ReportBase     string `json:"report_base"`
	Report         string `json:"report"`
	JobOrderContent
}

type JobOrderView struct {
	JobOrderID     *uint  `gorm:"column:id" json:"job_order_id"`
	BomId          uint   `json:"bom_id"`
	OrderDetailsId uint   `json:"order_details_id"`
	Date           string `json:"date"`
	SalesOrder     string `json:"sales_order"`
	Materials      string `json:"materials"`
	Quantity       int    `json:"quantity"`
	Due            string `json:"due"`
	A_Engr         string `json:"a_engr"`
	ItemRqst       string `json:"item_rqst"`
	Status         string `json:"status"`
	GeneralName    string `json:"general_name"`
	SerialNo       string `json:"serial_no"`
	ReportBase     string `json:"report_base"`
	Report         string `json:"report"`
	Type           string `json:"type"`
	ItemDesc       string `json:"item_desc"`
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

func (JobOrder) TableName() string {
	return "tbl_trans_job_order"
}

type JobOrderAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	JobOrderContent
	At
}

func (JobOrderAt) TableName() string {
	return "z_tbl_trans_job_order_at"
}
