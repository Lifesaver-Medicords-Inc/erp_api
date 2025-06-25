package accounting_models

import "github.com/pierceperado/smpc/models"

type SalesInvoiceDetailContent struct {
	SalesInvoiceId uint    `json:"sales_invoice_id"`
	OrderDetailsId uint    `json:"order_details_id"`
	BasedId        uint    `json:"based_id"`
	ItemId         uint    `json:"item_id"`
	ItemCode       string  `json:"item_code"`
	ItemDesc       string  `json:"item_desc"`
	Qty            uint    `json:"qty"`
	UnitPrice      float64 `json:"unit_price"`
	UomName        string  `json:"uom_name"`
	TotalCost      float64 `json:"total_cost"`
	Discount       float64 `json:"discount"`
	Status         string  `json:"status"`
	DeliveryDate   string  `json:"delivery_date"`
}

type SalesInvoiceDetail struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesInvoiceDetailContent
}

func (SalesInvoiceDetail) TableName() string {
	return "tbl_accounting_sales_invoice_details"
}

type SalesInvoiceDetailAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesInvoiceDetailContent
	models.At
}

func (SalesInvoiceDetailAt) TableName() string {
	return "z_tbl_accounting_sales_invoice_details_at"
}
