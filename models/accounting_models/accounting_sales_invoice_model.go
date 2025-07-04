package accounting_models

import "github.com/pierceperado/smpc/models"

type SalesInvoiceContent struct {
	CustomerId               uint    `json:"customer_id"`
	CustomerName             string  `json:"customer_name"`
	CustomerCode             string  `json:"customer_code"`
	PaymentTermsId           uint    `json:"payment_terms_id"`
	TaxCode                  string  `json:"tax_code"`
	JournalName              string  `json:"journal_name"`
	JournalId                uint    `json:"journal_id"`
	Address                  string  `json:"address"`
	SalesId                  string  `json:"sales_id"`
	SalesExecutive           string  `json:"sales_executive"`
	CurrencyCode             string  `json:"currency_code"`
	BaseRate                 uint    `json:"base_rate"`
	ExchangeRate             float64 `json:"exchange_rate"`
	DocNo                    string  `json:"doc_no"`
	DocDate                  string  `json:"doc_date"`
	PostingDate              string  `json:"posting_date"`
	ReferenceDrNo            string  `json:"reference_dr_no"`
	DocSoNo                  string  `json:"doc_so_no"`
	ReferencePo              string  `json:"reference_po"`
	VatableSales             float64 `json:"vatable_sales"`
	VatableSalesFc           float64 `json:"vatable_sales_fc"`
	VatExemptSales           float64 `json:"vat_exempt_sales"`
	VatExemptSalesFc         float64 `json:"vat_exempt_sales_fc"`
	ZeroRatedSales           float64 `json:"zero_rated_sales"`
	ZeroRatedSalesFc         float64 `json:"zero_rated_sales_fc"`
	TotalSalesVatInclusive   float64 `json:"total_sales_vat_inclusive"`
	TotalSalesVatInclusiveFc float64 `json:"total_sales_vat_inclusive_fc"`
	LessVat                  float64 `json:"less_vat"`
	LessVatFc                float64 `json:"less_vat_fc"`
	AmountNetVat             float64 `json:"amount_net_vat"`
	AmountNetVatFc           float64 `json:"amount_net_vat_fc"`
	LessPwdDiscount          float64 `json:"less_pwd_discount"`
	LessPwdDiscountFc        float64 `json:"less_pwd_discount_fc"`
	AmountDue                float64 `json:"amount_due"`
	AmountDueFc              float64 `json:"amount_due_fc"`
	AddVat                   float64 `json:"add_vat"`
	AddVatFc                 float64 `json:"add_vat_fc"`
	TotalAmountDue           float64 `json:"total_amount_due"`
	TotalAmountDueFc         float64 `json:"total_amount_due_fc"`
}
type SalesInvoice struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesInvoiceContent
}

func (SalesInvoice) TableName() string {
	return "tbl_accounting_sales_invoice"
}

type SalesInvoiceAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesInvoiceContent
	models.At
}

func (SalesInvoiceAt) TableName() string {
	return "z_tbl_sales_invoice_at"
}
