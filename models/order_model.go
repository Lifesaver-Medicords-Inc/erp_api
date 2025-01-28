package models

type OrderContent struct {
	Customer       string  `json:"customer"`
	Code           string  `json:"code"`
	DeliveryTo     string  `json:"delivery_to"`
	BillTo         string  `json:"bill_to"`
	DocumentNo     string  `json:"document_no"`
	Date           string  `json:"date"`
	DeliveryDate   string  `json:"delivery_date"`
	Payment_Terms  string  `json:"payment_terms"`
	Ship_Type      string  `json:"ship_type"`
	Ref_Doc        string  `json:"ref_doc"`
	Ref_Id         uint    `json:"ref_id"`
	Status         string  `json:"status"`
	SalesExecutive string  `json:"sales_executive"`
	Receiver       string  `json:"receiver"`
	ContactNo      string  `json:"contact_no"`
	Remarks        string  `json:"remarks"`
	Vat            float64 `json:"vat"`
	NetofVat       float64 `json:"net_of_vat"`
	TotalAmountDue float64 `json:"total_amount_due"`
	ApprovedBy     string  `json:"approved_by"`
	ApprovedByID   uint    `json:"approved_by_id"`
}

type Order struct {
	Order_ID uint `gorm:"primarykey" json:"order_id"`
	OrderContent
}

func (Order) TableName() string {
	return "tbl_trans_sales_order"
}

type OrderAt struct {
	Order_ID uint `gorm:"primarykey" json:"order_id"`
	RefId    uint `json:"ref_id"`
	OrderContent
	At
}

func (OrderAt) TableName() string {
	return "z_tbl_trans_sales_order_at"
}
