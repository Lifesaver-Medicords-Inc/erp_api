package accounting_models

import "github.com/pierceperado/smpc/models"

type SalesInvoiceContent2 struct {
	Customer        string  `json:"customer"`
	CustomerCode    string  `json:"customer_code"`
	CustomerId      uint    `json:"customer_id"`
	PaymentTerm     string  `json:"payment_term"`
	CustomerAddress string  `json:"customer_address"`
	TaxCode         string  `json:"tax_code"`
	Tin             string  `json:"tin"`
	Journal         string  `json:"journal"`
	BaseRate        float64 `json:"base_rate"`
	Currency        string  `json:"currency"`
	ExchangeRate    float64 `json:"exchange_rate"`
	DocNo           string  `json:"doc_no"`
	DocDate         string  `json:"doc_date"`
	PostingDate     string  `json:"posting_date"`
	ReferenceDocDr  string  `json:"reference_doc_dr"`
	ReferenceDocSo  string  `json:"reference_doc_so"`
	ReferenceDocPo  string  `json:"reference_doc_po"`
	SalesPerson     string  `json:"sales_person"`
	PreparedBy      string  `json:"prepared_by"`
	VatSales        float64 `json:"vat_sales"`
	VatExemptSales  float64 `json:"vat_exempt_sales"`
	ZeroSales       float64 `json:"zero_sales"`
	TotalSales      float64 `json:"total_sales"`
	LessVat         float64 `json:"less_vat"`
	NetVat          float64 `json:"net_vat"`
	PwdDiscount     float64 `json:"pwd_discount"`
	AmountDue       float64 `json:"amount_due"`
	AddVat          float64 `json:"add_vat"`
	TotalAmountDue  float64 `json:"total_amount_due"`
}

type SalesInvoice2 struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesInvoiceContent2
}

func (SalesInvoice2) TableName() string {
	return "tbl_accounting_sales_invoice2"
}

type SalesInvoice2At struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesInvoiceContent2
	models.At
}

func (SalesInvoice2At) TableName() string {
	return "z_tbl_accounting_sales_invoice2_at"
}

type SalesInvoiceDetails2Content struct {
	SalesInvoiceID           uint    `json:"sales_invoice_id"`
	SalesOrderDetailsId      uint    `json:"sales_order_details_id"`
	DeliveryReceiptDetailsId uint    `json:"delivery_receipt_details_id"`
	ItemId                   uint    `json:"item_id"`
	ItemCode                 string  `json:"item_code"`
	ItemDescription          string  `json:"item_description"`
	ItemQty                  uint    `json:"item_qty"`
	ItemUom                  string  `json:"item_uom"`
	UnitPrice                float64 `json:"unit_price"`
	Discount                 float64 `json:"discount"`
	TotalCost                float64 `json:"total_cost"`
	DateDeliver              string  `json:"date_deliver"`
}

type SalesInvoiceDetails2 struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesInvoiceDetails2Content
}

func (SalesInvoiceDetails2) TableName() string {
	return "tbl_accounting_sales_invoice2_details"
}

type SalesInvoiceDetails2At struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesInvoiceDetails2Content
	models.At
}

func (SalesInvoiceDetails2At) TableName() string {
	return "z_tbl_accounting_sales_invoice2_details_at"
}

type SalesInvoice2Body struct {
	SalesInvoice2        SalesInvoice2          `json:"sales_invoice2"`
	SalesInvoiceDetails2 []SalesInvoiceDetails2 `json:"sales_invoice_details2"`
}

type SalesInvoice2Get struct {
	SalesInvoice2        []SalesInvoice2        `json:"sales_invoice2"`
	SalesInvoiceDetails2 []SalesInvoiceDetails2 `json:"sales_invoice_details2"`
}
