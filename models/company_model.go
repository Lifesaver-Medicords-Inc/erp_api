package models

type CompanyContent struct {
	CompanyCode              string  `json:"company_code"`
	CompanyName              string  `json:"company_name"`
	LegalName                string  `json:"legal_name"`
	TradeName                string  `json:"trade_name"`
	BusinessType             string  `json:"business_type"`
	SecRegistrationNo        string  `json:"sec_registration_no"`
	DtiRegistrationNo        string  `json:"dti_registration_no"`
	Tin                      string  `json:"tin"`
	BirBranchCode            string  `json:"bir_branch_code"`
	RdoCode                  string  `json:"rdo_code"`
	Industry                 string  `json:"industry"`
	Status                   string  `json:"status"`
	IsHeadOffice             *bool   `json:"is_head_office"`
	BegBal                   float64 `json:"beg_bal"`
	MonthlyRate              float64 `json:"monthly_rate"`
	CurrencyCode             string  `json:"currency_code"`
	MarkUpMultiplierPrice    float64 `json:"markup_multiplier_price"`
	StartFiscalDate          string  `json:"start_fiscal_date"`
	EndFiscalDate            string  `json:"end_fiscal_date"`
	InclusionsQuotationTerms string  `json:"inclusions_quotation_terms"`
	ExclusionsQuotationTerms string  `json:"exclusions_quotation_terms"`
	TermAndConditions        string  `json:"term_and_conditions"`
}

type CompanyModel struct {
	ID uint `gorm:"primaryKey" json:"id"`
	CompanyContent
	Address  CompanyAddressModel   `gorm:"foreignKey:CompanyId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"address"`
	Contacts []CompanyContactModel `gorm:"foreignKey:CompanyId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"contacts"`
}

func (CompanyModel) TableName() string {
	return "tbl_company"
}

type CompanyAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	CompanyContent
	At
}

func (CompanyAt) TableName() string {
	return "z_tbl_company_at"
}

type CompanyCacheModel struct {
	ID                       uint    `gorm:"primaryKey" json:"id"`
	CompanyCode              string  `json:"company_code"`
	CompanyName              string  `json:"company_name"`
	LegalName                string  `json:"legal_name"`
	TradeName                string  `json:"trade_name"`
	BusinessType             string  `json:"business_type"`
	SecRegistrationNo        string  `json:"sec_registration_no"`
	DtiRegistrationNo        string  `json:"dti_registration_no"`
	Tin                      string  `json:"tin"`
	BirBranchCode            string  `json:"bir_branch_code"`
	RdoCode                  string  `json:"rdo_code"`
	Industry                 string  `json:"industry"`
	Status                   string  `json:"status"`
	IsHeadOffice             *bool   `json:"is_head_office"`
	BegBal                   float64 `json:"beg_bal"`
	MonthlyRate              float64 `json:"monthly_rate"`
	CurrencyCode             string  `json:"currency_code"`
	MarkUpMultiplierPrice    float64 `json:"markup_multiplier_price"`
	StartFiscalDate          string  `json:"start_fiscal_date"`
	EndFiscalDate            string  `json:"end_fiscal_date"`
	InclusionsQuotationTerms string  `json:"inclusions_quotation_terms"`
	ExclusionsQuotationTerms string  `json:"exclusions_quotation_terms"`
	TermAndConditions        string  `json:"term_and_conditions"`
}

func (CompanyCacheModel) TableName() string {
	return "tbl_company"
}

////-----

type CompanyAddressContent struct {
	AddressType  string `json:"address_type"`
	UnitNo       string `json:"unit_no"`
	BuildingName string `json:"building_name"`
	StreetName   string `json:"street_name"`
	Subdivision  string `json:"subdivision"`
	Barangay     string `json:"barangay"`
	City         string `json:"city"`
	Province     string `json:"province"`
	Region       string `json:"region"`
	Country      string `json:"country"`
	PostalCode   int    `json:"postal_code"`
}

type CompanyAddressModel struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	CompanyId uint `json:"company_id"`
	CompanyAddressContent
}

func (CompanyAddressModel) TableName() string {
	return "tbl_company_address"
}

type CompanyAddressAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	CompanyAddressContent
	At
}

func (CompanyAddressAt) TableName() string {
	return "z_tbl_company_address_at"
}

////-----

type CompanyContactContent struct {
	FullName         string `json:"full_name"`
	Designation      string `json:"designation"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phone_no"`
	MobileNumber     string `json:"mobile_no"`
	IsPrimaryContact *bool  `json:"is_primary_contact"`
}

type CompanyContactModel struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	CompanyId uint         `json:"company_id"`
	Company   CompanyModel `gorm:"foreignKey:CompanyId;references:ID;onDelete:CASCADE;onUpdate:CASCADE" json:"-"`
	CompanyContactContent
}

func (CompanyContactModel) TableName() string {
	return "tbl_company_contact"
}

type CompanyContactAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	CompanyContactContent
	At
}

func (CompanyContactAt) TableName() string {
	return "z_tbl_company_contact_at"
}
