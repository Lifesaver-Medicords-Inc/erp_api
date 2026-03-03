package accounting_models

type SalesInvoiceReceiptView struct {
	SalesInvoiceId int     `json:"sales_invoice_id"`
	DocNo          string  `json:"doc_no"`
	DocDate        string  `json:"doc_date"`
	DueDate        string  `json:"due_date"`
	TwasApplied    float64 `json:"twas_applied"`
	OpenAmount     float64 `json:"open_amount"`
}
