package models

type SalesQuotationContent struct {
	ProjectName          string  `json:"project_name"`
	CustomerID           uint    `json:"customer_id"`
	ApplicationID        uint    `json:"application_id"`
	PaymentTermsID       uint    `json:"payment_terms_id"`
	ShipToID             uint    `json:"ship_to_id"`
	BillToID             uint    `json:"bill_to_id"`
	ShipTypeID           uint    `json:"ship_type_id"`
	Purpose              string  `json:"purpose"`
	Date                 string  `json:"date"`
	ValidityDays         string  `json:"validity_days"`
	ValidUntil           string  `json:"valid_until"`
	Warranty             string  `json:"warranty"`
	AddressTo            string  `json:"address_to"`
	Thru                 string  `json:"thru"`
	GrossSales           float64 `json:"gross_sales"`
	VatAmount            float64 `json:"vat_amount"`
	NetSales             float64 `json:"net_sales"`
	PercentDiscount      float64 `json:"percent_discount"`
	DiscountedAmount     float64 `json:"discounted_amount"`
	AdditionalDiscounted float64 `json:"additional_discounted"`
	CashDiscount         float64 `json:"cash_discount"`
	NetAmountDue         float64 `json:"net_amount_due"`
	TotalAmountDue       float64 `json:"total_amount_due"`
	Contact1             string  `json:"contact_1" gorm:"column:contact_1"`
	Contact2             string  `json:"contact_2" gorm:"column:contact_2"`
	DocumentNo           string  `json:"document_no"`
	VersionNo            string  `json:"version_no"`
	SubVersionNo         string  `json:"sub_version_no"`
	CreatedBy            string  `json:"created_by"`
	FinalRefNo           string  `json:"final_ref_no"`
	IsFinalized          bool    `json:"is_finalized"`
	IsProject            bool    `json:"is_project"`

	// §3.2/§6.3 - REQUEST FOR ENGR. was a client-side stub (btn_request_for_engr_Click
	// just opened an old test form) with no backing field at all. The engineering red
	// box/Sales Quotation List instead inferred "sent to engineering" implicitly from
	// "has a project name and at least one wiring row" - which fires with no explicit
	// action and isn't scoped to any one engineer, contradicting §3.2's "an engineer
	// sees the quotations sent to them and no others" (a per-quote grant). These make
	// the grant explicit and assignable to a specific engineer - Phase 4 item 4.1.
	IsRequestedForEngr   bool   `json:"is_requested_for_engr"`
	RequestedEngrId      uint   `json:"requested_engr_id"`
	RequestedEngrName    string `json:"requested_engr_name"`
	RequestedForEngrDate string `json:"requested_for_engr_date"`
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
