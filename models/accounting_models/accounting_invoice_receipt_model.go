package accounting_models

import "github.com/pierceperado/smpc/models"

type InvoiceReceiptContent struct {
	Supplier        string  `json:"supplier"`
	SupplierCode    string  `json:"supplier_code"`
	SupplierId      uint    `json:"supplier_id"`
	SupplierAddress string  `json:"supplier_address"`
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
	ReferencePo     string  `json:"reference_po"`
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
	InvoiceReceiptID       uint    `json:"invoice_receipt_id"`
	ItemCode               string  `json:"item_code"`
	ItemDescription        string  `json:"item_description"`
	ReqQty                 uint    `json:"req_qty"`
	ReqUom                 string  `json:"req_uom"`
	UnitPrice              float64 `json:"unit_price"`
	TotalCost              float64 `json:"total_cost"`
	Discount               float64 `json:"discount"`
	LineAmount             float64 `json:"line_amount"`
	PurchaseOrderDetailsID uint    `json:"purchase_order_details_id"`
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
