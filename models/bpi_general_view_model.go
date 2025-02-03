package models

// Database View  for Bpi General
type BpiGeneralView struct {
	BpiGeneral
	BranchIndustryIds   string `json:"branch_industry_ids" `
	BranchIndustryNames string `json:"branch_industry_names"`
	EntityIds           string `json:"entity_ids"`
	EntityNames         string `json:"entity_names"`
}

func (BpiGeneralView) TableName() string {
	return "GetBpiGeneralList"
}
