package models

type BpiGeneralContent struct {
	BasedId         uint   `json:"based_id"`
	Social          uint   `json:"social_id"`
	BranchName      string `json:"branch_name"`
	TransactionType string `json:"transaction_type"`
	ClassName       string `json:"class_name"`
	BranchTelNo     string `json:"branch_tel_no"`
	BranchWebsite   string `json:"branch_website"`
	CustomerCode    string `json:"customer_code"`
	SupplierCode    string `json:"supplier_code"`
	FaxNo           string `json:"fax_no"`
	Notes           string `json:"notes"`
}

type BpiGeneral struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiGeneralContent
}

type BpiGeneralSchema struct {
	BpiGeneral
	BranchIndustriesId []uint `json:"branch_industry_id"`
	EntityTypeId       []uint `json:"entity_type_id"`
}

func (BpiGeneral) TableName() string {
	return "tbl_bpi_general"
}

type BpiGeneralAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiGeneralContent
	At
}

func (BpiGeneralAt) TableName() string {
	return "z_tbl_bpi_general_at"
}
