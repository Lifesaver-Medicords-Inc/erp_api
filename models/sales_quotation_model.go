package models

type SalesQuotationContent struct {
	CustomerName           string  `json:"customer_name"`
	CustomerCode           string  `json:"customer_code"`
	Name                   string  `json:"name"`
	Purpose                string  `json:"purpose"`
	Application            string  `json:"application"`
	PaymentTerms           string  `json:"payment_terms"`
	ShipType               string  `json:"ship_type"`
	ShipTo                 string  `json:"ship_to"`
	BillTo                 string  `json:"bill_to"`
	Date                   string  `json:"date"`
	ValidityDays           string  `json:"validity_days"`
	ValidUntil             string  `json:"valid_until"`
	Warranty               string  `json:"warranty"`
	AddressTo              string  `json:"address_to"`
	Thru                   string  `json:"thru"`
	GrossSales             float64 `json:"gross_sales"`
	VatAmount              float64 `json:"vat_amount"`
	NetSales               float64 `json:"net_sales"`
	SubTotalBeforeDiscount float64 `json:"sub_total_before_discount"`
	PercentDiscount        float64 `json:"percent_discount"`
	SubTotal               float64 `json:"sub_total"`
	CashDiscount           float64 `json:"cash_discount"`
	NetAmountDue           float64 `json:"net_amount_due"`
	IsVat                  bool    `json:"is_vat"`
	VatPercent             float64 `json:"vat_percent"`
	Contact1               string  `json:"contact_1"`
	Contact2               string  `json:"contact_2"`
	DocumentNo             string  `json:"document_no"`
	VersionNo              string  `json:"version_no"`
	CreatedBy              string  `json:"created_by"`
	DiscountedAmount       float64 `json:"discounted_amount"`
}

type SalesQuotation struct {
	ID             uint `gorm:"primarykey" json:"id"`
	CustomerID     uint `json:"customer_id"`
	ApplicationID  uint `json:"application_id"`
	PaymentTermsID uint `json:"payment_terms_id"`
	ShipToID       uint `json:"ship_to_id"`
	BillToID       uint `json:"bill_to_id"`
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
