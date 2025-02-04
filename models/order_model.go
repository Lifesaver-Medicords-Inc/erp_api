package models

type OrderContent struct {
	CustomerName   string `json:"customer_name"`
	CustomerCode   string `json:"customer_code"`
	ShipTo         string `json:"ship_to"`
	BillTo         string `json:"bill_to"`
	Doc            string `json:"doc"`
	Date           string `json:"date"`
	DeliveryDate   string `json:"delivery_date"`
	Payment_Terms  string `json:"payment_terms"`
	Ship_Type      string `json:"ship_type"`
	Document_No    string `json:"document_no"`
	Ref_Id         uint   `json:"ref_id"`
	Status         string `json:"status"`
	SalesExecutive string `json:"sales_executive"`
	Receiver       string `json:"receiver"`
	ContactNo      string `json:"contact_no"`
	Remarks        string `json:"remarks"`
	Vat            string `json:"vat"`
	NetofVat       string `json:"netof_vat"`
	TotalAmountDue string `json:"total_amount_due"`
	ApprovedBy     string `json:"approved_by"`
	ApprovedByID   uint   `json:"approved_by_id"`
	Tin            string `json:"tin"`
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
