package models

type SalesQuotationContent struct {
	ProjectName      string  `json:"project_name"`
	CustomerID       uint    `json:"customer_id"`
	ApplicationID    uint    `json:"application_id"`
	PaymentTermsID   uint    `json:"payment_terms_id"`
	ShipToID         uint    `json:"ship_to_id"`
	BillToID         uint    `json:"bill_to_id"`
	ShipTypeID       uint    `json:"ship_type_id"`
	Purpose          string  `json:"purpose"`
	Date             string  `json:"date"`
	ValidityDays     string  `json:"validity_days"`
	ValidUntil       string  `json:"valid_until"`
	Warranty         string  `json:"warranty"`
	AddressTo        string  `json:"address_to"`
	Thru             string  `json:"thru"`
	GrossSales       float64 `json:"gross_sales"`
	VatAmount        float64 `json:"vat_amount"`
	NetSales         float64 `json:"net_sales"`
	PercentDiscount  float64 `json:"percent_discount"`
	DiscountedAmount float64 `json:"discounted_amount"`
	CashDiscount     float64 `json:"cash_discount"`
	NetAmountDue     float64 `json:"net_amount_due"`
	TotalAmountDue   float64 `json:"total_amount_due"`
	Contact1         string  `json:"contact_1"`
	Contact2         string  `json:"contact_2"`
	DocumentNo       string  `json:"document_no"`
	VersionNo        string  `json:"version_no"`
	CreatedBy        string  `json:"created_by"`
	FinalRefNo       string  `json:"final_ref_no"`
}

type SalesQuotation struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesQuotationContent
}

func (SalesQuotation) TableName() string {
	return "tbl_trans_sales_quotation"
}

type SalesQuotationAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesQuotationContent
	At
}

func (SalesQuotationAt) TableName() string {
	return "z_tbl_trans_sales_quotation_at"
}
