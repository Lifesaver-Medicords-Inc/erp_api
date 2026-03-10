package accounting_models

import "github.com/pierceperado/smpc/models"

type BulkInvoiceReceiptContent struct {
	Supplier        string  `json:"supplier"`
	SupplierCode    string  `json:"supplier_code"`
	SupplierId      uint    `json:"supplier_id"`
	SupplierAddress string  `json:"supplier_address"`
	ReferenceDoc    string  `json:"reference_doc"`
	PaymentTerm     string  `json:"payment_term"`
	Currency        string  `json:"currency"`
	TaxCode         string  `json:"tax_code"`
	TaxCodeId       uint    `json:"tax_code_id"`
	InvoiceDue      string  `json:"invoice_due"`
	DocNo           string  `json:"doc_no"`
	DocDate         string  `json:"doc_date"`
	InvoiceType     string  `json:"invoice_type"`
	OtherCharges    float64 `json:"other_charges"`
	NetAmount       float64 `json:"net_amount"`
	TwasAmount      float64 `json:"twas_amount"`
	ApVoucher       *bool   `json:"ap_voucher"`
	Type            string  `json:"type"`
	Remarks         string  `json:"remarks"`
	PreparedBy      string  `json:"prepared_by"`
}

type BulkInvoiceReceipt struct {
	ID uint `gorm:"primarykey" json:"id"`
	BulkInvoiceReceiptContent
}

func (BulkInvoiceReceipt) TableName() string {
	return "tbl_accounting_bulk_invoice_receipt"
}

type BulkInvoiceReceiptAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BulkInvoiceReceiptContent
	models.At
}

func (BulkInvoiceReceiptAt) TableName() string {
	return "z_tbl_accounting_bulk_invoice_receipt_at"
}

type BulkInvoiceReceiptDetailsContent struct {
	BulkInvoiceReceiptID uint    `json:"bulk_invoice_receipt_id"`
	PaymentChargeCode    string  `json:"payment_charge_code"`
	ChargeDescription    string  `json:"charge_description"`
	AccountId            uint    `json:"account_id"`
	AccountCode          string  `json:"account_code"`
	LineAmount           float64 `json:"line_amount"`
}
type BulkInvoiceReceiptDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	BulkInvoiceReceiptDetailsContent
}

func (BulkInvoiceReceiptDetails) TableName() string {
	return "tbl_accounting_bulk_invoice_receipt_details"
}

type BulkInvoiceReceiptDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BulkInvoiceReceiptDetailsContent
	models.At
}

func (BulkInvoiceReceiptDetailsAt) TableName() string {
	return "z_tbl_accounting_bulk_invoice_receipt_details_at"
}

type BulkInvoiceReceiptBody struct {
	BulkInvoiceReceipt        BulkInvoiceReceipt          `json:"bulk_invoice_receipt"`
	BulkInvoiceReceiptDetails []BulkInvoiceReceiptDetails `json:"bulk_invoice_receipt_details"`
}

type BulkInvoiceReceiptGet struct {
	BulkInvoiceReceipt        []BulkInvoiceReceipt        `json:"bulk_invoice_receipt"`
	BulkInvoiceReceiptDetails []BulkInvoiceReceiptDetails `json:"bulk_invoice_receipt_details"`
}
