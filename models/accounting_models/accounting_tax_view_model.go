package accounting_models

type TaxView struct {
	Code             string  `json:"code"`
	TaxDesc          string  `json:"tax_desc"`
	OutputTaxAccount string  `json:"output_tax_account"`
	InputTaxAccount  string  `json:"input_tax_account"`
	TaxRate          float64 `json:"tax_rate"`
	EffectivePeriod  string  `json:"effective_period"`
}

func (TaxView) TableName() string {
	return "vw_get_tax_setup"
}
