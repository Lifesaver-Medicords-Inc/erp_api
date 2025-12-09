package models

type BpiGeneralEmbeddedContent struct {
	Social          uint   `json:"social_id"`
	TransactionType string `json:"transaction_type"`
	ClassName       string `json:"class_name"`
	BranchTelNo     string `json:"branch_tel_no"`
	BranchWebsite   string `json:"branch_website"`
	CustomerCode    string `json:"customer_code"`
	SupplierCode    string `json:"supplier_code"`
	FaxNo           string `json:"fax_no"`
	Notes           string `json:"notes"`
}

type BpiGeneralContent struct {
	BasedId    uint   `json:"based_id"`
	BranchName string `gorm:"unique;size:100" json:"branch_name"`
	SalesId    string `json:"sales_id"`
	IsMain     *bool  `json:"is_main"`
	BpiGeneralEmbeddedContent
}

type BpiGeneral struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiGeneralContent
}

type BpiGeneralSchema struct {
	BpiGeneral
	BranchIndustryId []uint `json:"branch_industry_id"`
	EntityTypeId     []uint `json:"entity_type_id"`
}

func (BpiGeneral) TableName() string {
	return "tbl_bpi_general"
}

type BpiGeneralAt struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	RefId      uint   `json:"ref_id"`
	BranchName string `gorm:"size:100" json:"branch_name"` // remove unique
	SalesId    string `json:"sales_id"`
	IsMain     *bool  `json:"is_main"`
	BpiGeneralEmbeddedContent
	At
}

func (BpiGeneralAt) TableName() string {
	return "z_tbl_bpi_general_at"
}
