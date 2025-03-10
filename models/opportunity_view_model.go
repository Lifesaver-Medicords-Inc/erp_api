package models

type OpportunityView struct {
	Tag         string `json:"tag"`
	VersionNo   string `json:"version_no"`
	FinalRefNo  string `json:"final_ref_no"`
	DocumentNo  string `json:"document_no"`
	Customer_ID uint   `json:"customer_id"`
	BranchName  string `json:"branch_name"`
	ProjectName string `json:"project_name"`
	Date        string `json:"date"`
	ClientReq   string `json:"client_req"`
	Value       string `json:"value"`
	LastUpdate  string `json:"last_update"`
	Stage       string `json:"stage"`
	Status      string `json:"status"`
	SpecialDeal string `json:"special_deal"`
	ID          uint   `json:"id"`
}

// test
func (OpportunityView) TableName() string {
	return "vw_get_sales_opportunities"
}
