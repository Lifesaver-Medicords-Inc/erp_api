package accounting_models

type InvoiceReceiptView struct {
	InvoiceReceiptId int     `json:"invoice_receipt_id"`
	ReceiptNo        string  `json:"receipt_no"`
	IrDocDate        string  `json:"ir_doc_date"`
	IrDueDate        string  `json:"ir_due_date"`
	TwasAmount       float64 `json:"twas_amount"`
	LineAmount       float64 `json:"line_amount"`
	ReceiptType      string  `json:"receipt_type"`

	// Real running balance (net_amount minus everything already applied via
	// any AP Voucher or Debit Memo) - see services.ComputeReceiptOpenAmount.
	// Previously this view had no such column at all; AP Voucher and Debit
	// Memo's shared picker (sp_GetInvoiceAPVoucher) both just echoed
	// LineAmount and filtered on the all-or-nothing ap_voucher boolean,
	// which is why neither document could ever partially apply against one
	// of these.
	OpenAmount float64 `json:"open_amount"`
}
