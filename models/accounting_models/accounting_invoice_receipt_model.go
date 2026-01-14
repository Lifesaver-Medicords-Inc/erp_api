package accounting_models

import "github.com/pierceperado/smpc/models"

type InvoiceReceiptContent struct {
	Supplier     string `json:"supplier"`
	SupplierCode string `json:"supplier_code"`
	PaymentTerm  string `json:"payment_term"`
	TaxCode      string `json:"tax_code"`
	InvoiceDue   string `json:"invoice_due"`
	DocNo        string `json:"doc_no"`
	DocDate      string `json:"doc_date"`
	InvoiceType  string `json:"invoice_type"`
	NetAmount    string `json:"net_amount"`
	TwasAmount   string `json:"twas_amount"`
	ApVoucher    string `json:"ap_voucher"`
	Type         string `json:"type"`
	Remarks      string `json:"remarks"`
}
type InvoiceReceipt struct {
	ID uint `gorm:"primarykey" json:"id"`
	InvoiceReceiptContent
}

func (InvoiceReceipt) TableName() string {
	return "tbl_accounting_invoice_receipt"
}

type InvoiceReceiptAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	InvoiceReceiptContent
	models.At
}

func (InvoiceReceiptAt) TableName() string {
	return "z_tbl_accounting_invoice_receipt_at"
}

type InvoiceReceiptDetailsContent struct {
	InvoiceReceiptID uint    `json:"invoice_receipt_id"`
	ItemCode         string  `json:"item_code"`
	ItemDescription  string  `json:"item_description"`
	ItemQty          uint    `json:"item_qty"`
	UnitPrice        float64 `json:"unit_price"`
	TotalCost        float64 `json:"total_cost"`
	Discount         float64 `json:"discount"`
	LineAmount       float64 `json:"line_amount"`
}
type InvoiceReceiptDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	InvoiceReceiptDetailsContent
}

func (InvoiceReceiptDetails) TableName() string {
	return "tbl_accounting_invoice_receipt_details"
}

type InvoiceReceiptDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	InvoiceReceiptDetailsContent
	models.At
}

func (InvoiceReceiptDetailsAt) TableName() string {
	return "z_tbl_accounting_invoice_receipt_details_at"
}

type InvoiceReceiptBody struct {
	InvoiceReceipt        InvoiceReceipt          `json:"invoice_receipt"`
	InvoiceReceiptDetails []InvoiceReceiptDetails `json:"invoice_receipt_details"`
}

type InvoiceReceiptGet struct {
	InvoiceReceipt        []InvoiceReceipt        `json:"invoice_receipt"`
	InvoiceReceiptDetails []InvoiceReceiptDetails `json:"invoice_receipt_details"`
}
