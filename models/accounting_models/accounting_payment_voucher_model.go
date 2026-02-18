package accounting_models

import "github.com/pierceperado/smpc/models"

type PaymentVoucherContent struct {
	Supplier          string  `json:"supplier"`
	SupplierCode      string  `json:"supplier_code"`
	SupplierId        uint    `json:"supplier_id"`
	ReferenceApv      string  `json:"reference_apv"`
	Currency          string  `json:"currency"`
	TransactionAmount float64 `json:"transaction_amount"`
	DocNo             string  `json:"doc_no"`
	DocDate           string  `json:"doc_date"`
	Remarks           string  `json:"remarks"`
	CashAmount        float64 `json:"cash_amount"`
	CheckAmount       float64 `json:"check_amount"`
	CheckBank         string  `json:"check_bank"`
	CheckAccountNo    string  `json:"check_account_no"`
	RefCheckNo        string  `json:"ref_check_no"`
	CheckDate         string  `json:"check_date"`
	TransferAmount    float64 `json:"transfer_amount"`
	TransferType      string  `json:"transfer_type"`
	TransferBank      string  `json:"transfer_bank"`
	TransferAccountNo string  `json:"transfer_account_no"`
	RefDocNo          string  `json:"ref_doc_no"`
	RefDocDate        string  `json:"ref_doc_date"`
	PreparedBy        string  `json:"prepared_by"`
}

type PaymentVoucher struct {
	ID uint `gorm:"primarykey" json:"id"`
	PaymentVoucherContent
}

func (PaymentVoucher) TableName() string {
	return "tbl_accounting_payment_voucher"
}

type PaymentVoucherAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PaymentVoucherContent
	models.At
}

func (PaymentVoucherAt) TableName() string {
	return "z_tbl_accounting_payment_voucher_at"
}

type PaymentVoucherDetailsContent struct {
	PaymentVoucherID   uint    `json:"payment_voucher_id"`
	ApVoucherDetailsId uint    `json:"ap_voucher_details_id"`
	DocNo              string  `json:"doc_no"`
	DueDate            string  `json:"due_date"`
	TransAmount        float64 `json:"trans_amount"`
	OpenAmount         float64 `json:"open_amount"`
	AmountApplied      float64 `json:"amount_applied"`
	TwasApplied        float64 `json:"twas_applied"`
	Balance            float64 `json:"balance"`
}
type PaymentVoucherDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	PaymentVoucherDetailsContent
}

func (PaymentVoucherDetails) TableName() string {
	return "tbl_accounting_payment_voucher_details"
}

type PaymentVoucherDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PaymentVoucherDetailsContent
	models.At
}

func (PaymentVoucherDetailsAt) TableName() string {
	return "z_tbl_accounting_payment_voucher_details_at"
}

type PaymentVoucherBody struct {
	PaymentVoucher        PaymentVoucher          `json:"payment_voucher"`
	PaymentVoucherDetails []PaymentVoucherDetails `json:"payment_voucher_details"`
}

type PaymentVoucherGet struct {
	PaymentVoucher        []PaymentVoucher        `json:"payment_voucher"`
	PaymentVoucherDetails []PaymentVoucherDetails `json:"payment_voucher_details"`
}
