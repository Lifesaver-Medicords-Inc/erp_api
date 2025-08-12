package accounting_models

type TaxClassification struct {
	Id                 uint    `json:"Id"`
	Code               string  `json:"code"`
	InputTaxCreditable *bool   `json:"input_tax_creditable"`
	TaxDesc            string  `json:"tax_desc"`
	TaxRate            float64 `json:"tax_rate"`
	AccountId          uint    `json:"account_id"`
	Status             string  `json:"status"`
}
