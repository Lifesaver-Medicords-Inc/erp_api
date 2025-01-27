package models

type OrderContent struct {
	Based_ID           uint   `json:"based_id"`
	Qty                string `json:"qty"`
	ItemCode           string `json:"item_code"`
	ItemDescription    string `json:"item_description"`
	DeliveryPreference string `json:"delivery_preference"`
	ListPrice          string `json:"list_price"`
	TotalPrice         string `json:"total_price"`
	Status             string `json:"status"`
	Receiver           string `json:"receiver"`
	ContactNo          string `json:"contact_no"`
	Remarks            string `json:"remarks"`
	Vat                string `json:"vat"`
	NetofVat           string `json:"net_of_vat"`
	TotalAmountDue     string `json:"total_amount_due"`
	ApprovedBy         string `json:"approved_by"`
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
