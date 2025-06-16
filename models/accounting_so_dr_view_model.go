package models

type SalesOrderWithDeliveryReceipt struct {
	OrderId        uint    `json:"order_id"`
	CustomerId     uint    `json:"customer_id"`
	PaymentTermsId uint    `json:"payment_terms_id"`
	RefPo          uint    `json:"ref_po"`
	CustomerName   string  `json:"customer_name"`
	CustomerCode   string  `json:"customer_code"`
	TaxCode        string  `json:"tax_code"`
	Tax            string  `json:"tax"`
	Address        string  `json:"address"`
	DocSoNo        string  `json:"doc_so_no"`
	DocDate        string  `json:"doc_date"`
	NetAmount      float64 `json:"net_amount"`
	SalesExecutive string  `json:"sales_executive"`
}

func (SalesOrderWithDeliveryReceipt) TableName() string {
	return "vw_get_sales_order_with_delivery_receipts"
}
