package models

type SalesOrderWithDRDetails struct {
	OrderDetailsId uint    `json:"order_details_id"`
	BasedId        uint    `json:"based_id"`
	ItemId         uint    `json:"item_id"`
	ItemCode       string  `json:"item_code"`
	ItemDesc       string  `json:"item_desc"`
	Qty            uint    `json:"qty"`
	UnitPrice      float64 `json:"unit_price"`
	TotalCost      float64 `json:"total_cost"`
	Discount       uint    `json:"discount"`
	Status         string  `json:"status"`
	DeliveryDate   string  `json:"delivery_date"`
}

func (SalesOrderWithDRDetails) TableName() string {
	return "vw_get_sales_order_dr_details"
}
