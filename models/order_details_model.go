package models

type OrderDetailsContent struct {
	Based_ID           uint    `json:"based_id"`
	Quotation_Quick_ID uint    `json:"quotation_quick_id"`
	Item_ID            uint    `json:"item_id"`
	DeliveryPreference string  `json:"delivery_preference"`
	Status             string  `json:"status"`
	HasStocks          bool    `json:"has_stocks"`
	Qty                int     `json:"qty"`
	ItemCode           string  `json:"item_code"`
	ItemDescription    string  `json:"item_description"`
	ListPrice          float64 `json:"list_price"`
	TotalPrice         float64 `json:"total_price"`
}

type OrderDetails struct {
	Order_Details_ID uint `gorm:"primarykey" json:"order_details_id"`
	OrderDetailsContent
}

func (OrderDetails) TableName() string {
	return "tbl_trans_sales_order_details"
}

type OrderDetailsAt struct {
	Order_Details_ID uint `gorm:"primarykey" json:"order_details_id"`
	RefId            uint `json:"ref_id"`
	OrderDetailsContent
	At
}

func (OrderDetailsAt) TableName() string {
	return "z_tbl_trans_sales_order_details_at"
}
