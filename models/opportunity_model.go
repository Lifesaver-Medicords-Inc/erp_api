package models

type OpportunityContent struct {
	TAG         string `json:"tag"`
	ClientReq   string `json:"client_req"`
	LastUpdate  string `json:"last_update"`
	Stage       string `json:"stage"`
	Status      string `json:"status"`
	SpecialDeal string `json:"special_deal"`
	DocumentNo  string `json:"document_no"`
}

type Opportunity struct {
	ID uint `gorm:"primarykey" json:"id"`
	OpportunityContent
}

func (Opportunity) TableName() string {
	return "tbl_trans_sales_opportunity"
}

type OpportunityAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	OpportunityContent
	At
}

// testsadas
func (OpportunityAt) TableName() string {
	return "z_tbl_trans_sales_opportunity_at"
}
