package models

type ItemSalesViewContent struct {
	BasedId      uint   `json:"based_id"`
	CustomerId   uint   `json:"customer_id"`
	SalesOrderNo string `json:"sales_order_no"`
	Date         string `json:"date"`
	CustomerName string `json:"customer_name"`
}

type ItemSalesView struct{
	ID uint `gorm:"primarykey" json:"id"`
	ItemSalesViewContent
}

func (ItemSalesView) TableName() string {
	return "vw_item_sales_list"
}
