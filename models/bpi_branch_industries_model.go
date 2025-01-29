package models

type BpiBranchIndustriesContent struct {
	BpiGeneralId uint `json:"bpi_general_id"`
	IndustryId   uint `json:"industry_id"`
}

type BpiBranchIndustries struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiBranchIndustriesContent
}

func (BpiBranchIndustries) TableName() string {
	return "tbl_bpi_branch_industries"
}

type BpiBranchIndustriesAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiBranchIndustriesContent
	At
}

func (BpiBranchIndustriesAt) TableName() string {
	return "z_tbl_bpi_branch_industries_at"
}
