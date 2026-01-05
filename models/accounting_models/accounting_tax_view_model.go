package accounting_models

type TaxView struct {
	ViewId           int     `json:"view_id"`
	Code             string  `json:"code"`
	TaxDesc          string  `json:"tax_desc"`
	CoaPurchId       int     `json:"coa_purch_id"`
	OutputTaxAccount string  `json:"output_tax_account"`
	CoaSalesId       int     `json:"coa_sales_id"`
	InputTaxAccount  string  `json:"input_tax_account"`
	TaxRate          float64 `json:"tax_rate"`
	EffectivePeriod  string  `json:"effective_period"`
	AccountType      string  `json:"account_type"`
}

func (TaxView) TableName() string {
	return "vw_get_tax_setup"
}

type CoaView struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (CoaView) TableName() string {
	return "vw_get_coa_setup"
}
