package accounting_models

type SupplierTradeView struct {
	SupplierID        string  `json:"supplier_id"`
	Supplier          string  `json:"supplier"`
	SupplierCode      string  `json:"supplier_code"`
	SupplierAddress   string  `json:"supplier_address"`
	InvoiceType       string  `json:"invoice_type"`
	PaymentTerm       string  `json:"payment_term"`
	Type              string  `json:"type"`
	OverpaymentAmount float64 `json:"overpayment_amount"`
}

func (SupplierTradeView) TableName() string {
	return "vw_get_supplier_trade"
}
