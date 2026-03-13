package accounting_models

import "github.com/pierceperado/smpc/models"

type PaymentReceiptContent struct {
	Customer          string  `json:"customer"`
	CustomerCode      string  `json:"customer_code"`
	CustomerId        uint    `json:"customer_id"`
	CustomerAddress   string  `json:"customer_address"`
	ReferenceCrNo     string  `json:"reference_cr_no"`
	DateCollect       string  `json:"date_collect"`
	ReferenceOrNo     string  `json:"reference_or_no"`
	CashAmount        float64 `json:"cash_amount"`
	TransactionAmount float64 `json:"transaction_amount"`
	Currency          string  `json:"currency"`
	UnappliedAmount   float64 `json:"unapplied_amount"`
	DocNo             int     `json:"doc_no"`
	DocDate           string  `json:"doc_date"`
	BankCode          string  `json:"bank_code"`
	BankBranch        string  `json:"bank_branch"`
	CheckNo           string  `json:"check_no"`
	CheckType         string  `json:"check_type"`
	CheckDate         string  `json:"check_date"`
	CheckAmount       float64 `json:"check_amount"`
	TransferAmount    float64 `json:"transfer_amount"`
	TransferType      string  `json:"transfer_type"`
	TransferBank      string  `json:"transfer_bank"`
	TransferAccountNo string  `json:"transfer_account_no"`
	RefDocNo          string  `json:"ref_doc_no"`
	RefDocDate        string  `json:"ref_doc_date"`
	PreparedBy        string  `json:"prepared_by"`
}

type PaymentReceipt struct {
	ID uint `gorm:"primarykey" json:"id"`
	PaymentReceiptContent
}

func (PaymentReceipt) TableName() string {
	return "tbl_accounting_payment_receipt"
}

type PaymentReceiptAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PaymentReceiptContent
	models.At
}

func (PaymentReceiptAt) TableName() string {
	return "z_tbl_accounting_payment_receipt_at"
}

type PaymentReceiptDetailsContent struct {
	PaymentReceiptID      uint    `json:"payment_receipt_id"`
	SalesOrderDetailsId   uint    `json:"sales_order_details_id"`
	SalesInvoiceDetailsId uint    `json:"sales_invoice_details_id"`
	SalesInvoiceId        uint    `json:"sales_invoice_id"`
	DocNo                 int     `json:"doc_no"`
	DocDate               string  `json:"doc_date"`
	DueDate               string  `json:"due_date"`
	OpenAmount            float64 `json:"open_amount"`
	AmountApplied         float64 `json:"amount_applied"`
	TwasApplied           float64 `json:"twas_applied"`
	Balance               float64 `json:"balance"`
}

type PaymentReceiptDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	PaymentReceiptDetailsContent
}

func (PaymentReceiptDetails) TableName() string {
	return "tbl_accounting_payment_receipt_details"
}

type PaymentReceiptDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PaymentReceiptDetailsContent
	models.At
}

func (PaymentReceiptDetailsAt) TableName() string {
	return "z_tbl_accounting_payment_receipt_details_at"
}

type PaymentReceiptBody struct {
	PaymentReceipt        PaymentReceipt          `json:"payment_receipt"`
	PaymentReceiptDetails []PaymentReceiptDetails `json:"payment_receipt_details"`
}

type PaymentReceiptGet struct {
	PaymentReceipt        []PaymentReceipt        `json:"payment_receipt"`
	PaymentReceiptDetails []PaymentReceiptDetails `json:"payment_receipt_details"`
}
