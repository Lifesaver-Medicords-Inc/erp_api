package models

type OrderContent struct {
	Ship_Type_ID     uint   `json:"ship_type_id"`
	Bill_To_ID       uint   `json:"bill_to_id"`
	Ship_To_ID       uint   `json:"ship_to_id"`
	Customer_ID      uint   `json:"customer_id"`
	Quotation_ID     uint   `json:"quotation_id"`
	Payment_Terms_ID uint   `json:"payment_terms_id"`
	ApprovedBy       string `json:"approved_by"`
	ApprovedByID     uint   `json:"approved_by_id"`
	Doc              uint   `json:"doc"`
	RefPO            uint   `json:"ref_po"`
	Date             string `json:"date"`
	DeliveryDate     string `json:"delivery_date"`
	Document_No      uint   `json:"document_no"`
	Status           string `json:"status"`
	// CustomerName  string `json:"customer_name"`
	// CustomerCode  string `json:"customer_code"`
	// ShipTo        string `json:"ship_to"`
	// BillTo        string `json:"bill_to"`
	// SalesExecutive string `json:"sales_executive"`
	// Receiver       string `json:"receiver"`
	// ContactNo      string `json:"contact_no"`
	// Remarks        string `json:"remarks"`
	// Vat            string `json:"vat"`
	// NetofVat       string `json:"netof_vat"`
	// TotalAmountDue string `json:"total_amount_due"`
	// ApprovedBy     string `json:"approved_by"`
	// ApprovedByID   uint   `json:"approved_by_id"`
	// Tin            string `json:"tin"`
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
