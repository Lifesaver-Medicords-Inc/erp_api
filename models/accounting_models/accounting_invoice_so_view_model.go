package accounting_models

type InvoiceSOView struct {
	SalesOrderID uint    `json:"sales_order_id"`
	SONumber     string  `json:"so_number"`
	DocDate      string  `json:"doc_date"`
	CustomerName string  `json:"customer_name"`
	SalesPerson  string  `json:"sales_person"`
	TotalSales   float64 `json:"total_sales"`
}

type InvoiceSODetailView struct {
	SalesOrderDetailsId uint    `json:"sales_order_details_id"`
	SalesOrderId        uint    `json:"sales_order_id"`
	ItemId              uint    `json:"item_id"`
	ItemCode            string  `json:"item_code"`
	ItemDescription     string  `json:"item_description"`
	ItemUom             string  `json:"item_uom"`
	ItemQty             uint    `json:"item_qty"`
	Discount            float64 `json:"discount"`
	UnitPrice           float64 `json:"unit_price"`
	TotalCost           float64 `json:"total_cost"`
	DateDeliver         string  `json:"date_deliver"`
}
