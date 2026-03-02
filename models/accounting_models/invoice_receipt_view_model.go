package accounting_models

type InvoiceReceiptView struct {
	InvoiceReceiptId int     `json:"invoice_receipt_id"`
	ReceiptNo        string  `json:"receipt_no"`
	IrDocDate        string  `json:"ir_doc_date"`
	IrDueDate        string  `json:"ir_due_date"`
	TwasAmount       float64 `json:"twas_amount"`
	LineAmount       float64 `json:"line_amount"`
	ReceiptType      string  `json:"receipt_type"`
}
