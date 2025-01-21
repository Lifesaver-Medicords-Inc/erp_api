package models

type SalesQuotation struct {
	ID                     uint   `gorm:"primarykey" json:"id"`
	CustomerName           string `json:"customer_name"`
	CustomerCode           string `json:"customer_code"`
	Purpose                string `json:"purpose"`
	Application            string `json:"application"`
	PaymentTerms           string `json:"payment_terms"`
	ShipType               string `json:"ship_type"`
	ShipTo                 string `json:"ship_to"`
	BillTo                 string `json:"bill_to"`
	Date                   string `json:"date"`
	ValidityDays           string `json:"validity_days"`
	ValidUntil             string `json:"valid_until"`
	Warranty               string `json:"warranty"`
	AddressTo              string `json:"address_to"`
	Thru                   string `json:"thru"`
	GrossSales             string `json:"gross_sales"`
	VatAmount              string `json:"vat_amount"`
	NetSales               string `json:"net_sales"`
	SubTotalBeforeDiscount string `json:"sub_total_before_discount"`
	PercentDiscount        string `json:"percent_discount"`
	SubTotal               string `json:"sub_total"`
	CashDiscount           string `json:"cash_discount"`
	NetAmountDue           string `json:"net_amount_due"`
	IsVat                  string `json:"isVat"`
	VatPercent             string `json:"vat_percent"`
	Contact1               string `json:"contact_1"`
	Contact2               string `json:"contact_2"`
	DocumentNo             string `json:"document_no"`
	VersionNo              string `json:"version_no"`
	CreatedBy              string `json:"created_by"`
	DiscountedAmount       string `json:"discounted_amount"`
}

func (SalesQuotation) TableName() string {
	return "tbl_trans_sales_quotation"
}

type SalesQuotationAt struct {
	ID                     uint   `gorm:"primarykey" json:"id"`
	RefId                  uint   `json:"ref_id"`
	CustomerName           string `json:"customer_name"`
	CustomerCode           string `json:"customer_code"`
	Purpose                string `json:"purpose"`
	Application            string `json:"application"`
	PaymentTerms           string `json:"payment_terms"`
	ShipType               string `json:"ship_type"`
	ShipTo                 string `json:"ship_to"`
	BillTo                 string `json:"bill_to"`
	Date                   string `json:"date"`
	ValidityDays           string `json:"validity_days"`
	ValidUntil             string `json:"valid_until"`
	Warranty               string `json:"warranty"`
	AddressTo              string `json:"address_to"`
	Thru                   string `json:"thru"`
	GrossSales             string `json:"gross_sales"`
	VatAmount              string `json:"vat_amount"`
	NetSales               string `json:"net_sales"`
	SubTotalBeforeDiscount string `json:"sub_total_before_discount"`
	PercentDiscount        string `json:"percent_discount"`
	SubTotal               string `json:"sub_total"`
	CashDiscount           string `json:"cash_discount"`
	NetAmountDue           string `json:"net_amount_due"`
	IsVat                  string `json:"isVat"`
	VatPercent             string `json:"vat_percent"`
	Contact1               string `json:"contact_1"`
	Contact2               string `json:"contact_2"`
	DocumentNo             string `json:"document_no"`
	VersionNo              string `json:"version_no"`
	CreatedBy              string `json:"created_by"`
	DiscountedAmount       string `json:"discounted_amount"`
	At                     `json:"at"`
}

func (SalesQuotationAt) TableName() string {
	return "z_tbl_trans_sales_quotation_at"
}
