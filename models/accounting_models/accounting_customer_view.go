package accounting_models

type CustomerView struct {
	CustomerID        string  `json:"customer_id"`
	Customer          string  `json:"customer"`
	CustomerCode      string  `json:"customer_code"`
	PaymentTerm       string  `json:"payment_term"`
	TaxCode           string  `json:"tax_code"`
	Tax               string  `json:"tax"`
	CustomerAddress   string  `json:"customer_address"`
	OverpaymentAmount float64 `json:"overpayment_amount"`
}

func (CustomerView) TableName() string {
	return "vw_get_customer"
}
