package models

// Database View  for Bpi General
type BpiGeneralView struct {
	GeneralId      uint   `json:"general_id"`
	GeneralBasedId uint   `json:"general_based_id"`
	BranchSalesId  string `json:"branch_sales_id"`
	BranchName     string `json:"branch_name"`
	IsMain         bool   `json:"is_main"`
	BpiGeneralEmbeddedContent
	BranchIndustryIds   string `json:"branch_industry_ids" `
	BranchIndustryNames string `json:"branch_industry_names"`
	EntityIds           string `json:"entity_ids"`
	EntityNames         string `json:"entity_names"`
}

func (BpiGeneralView) TableName() string {
	return "GetBpiGeneralList"
}
