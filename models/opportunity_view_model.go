package models

type OpportunityView struct {
	QuoteRef     string `json:"quote_ref"`
	Tag          string `json:"tag"`
	DocumentNo   string `json:"document_no"`
	CustomerName string `json:"customer_name"`
	ProjectName  string `json:"project_name"`
	Date         string `json:"date"`
	ClientReq    string `json:"client_req"`
	Value        string `json:"value"`
	LastUpdate   string `json:"last_update"`
	Stage        string `json:"stage"`
	Status       string `json:"status"`
	SpecialDeal  string `json:"special_deal"`
}

func (OpportunityView) TableName() string {

	return "vw_sales_opportunities"
}
