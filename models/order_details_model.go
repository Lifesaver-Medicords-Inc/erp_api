package models

type OrderDetailsContent struct {
	Based_ID           uint    `json:"based_id"`
	Quotation_Quick_ID uint    `json:"quotation_quick_id"`
	Item_ID            uint    `json:"item_id"`
	DeliveryPreference string  `json:"delivery_preference"`
	Status             string  `json:"status"`
	HasStocks          *bool   `json:"has_stocks"`
	Qty                *int    `json:"qty"`
	Numbering          string  `json:"numbering"`
	ItemCode           string  `json:"item_code"`
	ItemDescription    string  `json:"item_description"`
	ListPrice          float64 `json:"list_price"`
	PercentDiscount    float32 `json:"percent_discount"`
	TotalPrice         float64 `json:"total_price"`
	AllocatedQty       *int    `json:"allocated_qty"`
	OrderType          string  `json:"order_type"`
	BomId              uint    `json:"bom_id"`
	Item               *Item   `json:"item"`
	Order              *Order  `gorm:"foreignKey:Based_ID;references:Order_ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order,omitempty"`
}

type OrderDetails struct {
	OrderDetailsID uint `gorm:"primarykey" json:"order_details_id"`
	OrderDetailsContent
}

func (OrderDetails) TableName() string {
	return "tbl_trans_sales_order_details"
}

type OrderDetailsAt struct {
	OrderDetailsID uint `gorm:"primarykey" json:"order_details_id"`
	RefId          uint `json:"ref_id"`
	OrderDetailsContent
	At
}

func (OrderDetailsAt) TableName() string {
	return "z_tbl_trans_sales_order_details_at"
}
