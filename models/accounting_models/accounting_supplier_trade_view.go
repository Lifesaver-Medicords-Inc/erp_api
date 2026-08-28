package accounting_models

type SupplierTradeView struct {
	SupplierID        string  `json:"supplier_id"`
	Supplier          string  `json:"supplier"`
	SupplierCode      string  `json:"supplier_code"`
	SupplierAddress   string  `json:"supplier_address"`
	InvoiceType       string  `json:"invoice_type"`
	PaymentTerm       string  `json:"payment_term"`
	// TaxCode is the BPI's own configured code (e.g. "S1") from
	// tbl_bpi_finance.finance_tax_code - a plain string, not a Tax Setup id -
	// InvoiceReceiptPage.cs looks this up against its Tax Setup combo by code.
	TaxCode           string  `json:"tax_code"`
	Type              string  `json:"type"`
	OverpaymentAmount float64 `json:"overpayment_amount"`
}

func (SupplierTradeView) TableName() string {
	return "vw_get_supplier_trade"
}
