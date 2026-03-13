package accounting_models

import (
	"github.com/pierceperado/smpc/models"
)

type ApVoucherContent struct {
	Supplier          string  `json:"supplier"`
	SupplierCode      string  `json:"supplier_code"`
	SupplierId        uint    `json:"supplier_id"`
	Currency          string  `json:"currency"`
	DocNo             int     `json:"doc_no"`
	DocDate           string  `json:"doc_date"`
	TransactionAmount float64 `json:"transaction_amount"`
	PreparedBy        string  `json:"prepared_by"`
}

type ApVoucher struct {
	ID uint `gorm:"primarykey" json:"id"`
	ApVoucherContent
}

func (ApVoucher) TableName() string {
	return "tbl_accounting_ap_voucher"
}

type ApVoucherAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ApVoucherContent
	models.At
}

func (ApVoucherAt) TableName() string {
	return "z_tbl_accounting_ap_voucher_at"
}

type ApVoucherDetailsContent struct {
	ApVoucherId      uint    `json:"ap_voucher_id"`
	InvoiceReceiptId uint    `json:"invoice_receipt_id"`
	ReceiptNo        int     `json:"receipt_no"`
	IrDocDate        string  `json:"ir_doc_date"`
	IrDueDate        string  `json:"ir_due_date"`
	TwasAmount       float64 `json:"twas_amount"`
	LineAmount       float64 `json:"line_amount"`
	ReceiptType      string  `json:"receipt_type"`
}

type ApVoucherDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	ApVoucherDetailsContent
}

func (ApVoucherDetails) TableName() string {
	return "tbl_accounting_ap_voucher_details"
}

type ApVoucherDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ApVoucherDetailsContent
	models.At
}

func (ApVoucherDetailsAt) TableName() string {
	return "z_tbl_accounting_ap_voucher_details_at"
}

type ApVoucherBody struct {
	ApVoucher        ApVoucher          `json:"ap_voucher"`
	ApVoucherDetails []ApVoucherDetails `json:"ap_voucher_details"`
}

type ApVoucherGet struct {
	ApVoucher        []ApVoucher        `json:"ap_voucher"`
	ApVoucherDetails []ApVoucherDetails `json:"ap_voucher_details"`
}
