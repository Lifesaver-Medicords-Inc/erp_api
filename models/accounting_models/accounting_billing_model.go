package accounting_models

import "github.com/pierceperado/smpc/models"

// The SO's running billing ledger (spec 12.4) - hand-keyed by accounting, and
// the only place a deposit appears. It is NOT the down-payment feature (spec 15
// keeps that out of scope): a CLIENT PAY row records a payment, it does not
// raise a receipt or carry an unapplied advance.
//
// Amounts are signed the way the ledger reads: BEGINING BALANCE and ADDITIONAL
// are positive, CLIENT PAY is negative. The billing list sums them as they are.
type SoBillingTransactionContent struct {
	SoNo            string  `json:"so_no"`
	Date            string  `json:"date"`
	TransactionType string  `json:"transaction_type"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Reference       string  `json:"reference"`
	Remarks         string  `json:"remarks"`
	CreatedBy       string  `json:"created_by"`
}

type SoBillingTransaction struct {
	ID uint `gorm:"primarykey" json:"id"`
	SoBillingTransactionContent
}

func (SoBillingTransaction) TableName() string {
	return "tbl_accounting_so_billing_transaction"
}

type SoBillingTransactionAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SoBillingTransactionContent
	models.At
}

func (SoBillingTransactionAt) TableName() string {
	return "z_tbl_accounting_so_billing_transaction_at"
}

// ---------------------------------------------------------------- read models

// One row of the A/R Record of Transactions (spec 12.3): one per Sales Invoice.
// NEXT DUE is typed by A/R - no setup converts a payment term into a date - and
// is stored on the invoice.
type ARRecordRow struct {
	SalesInvoiceId uint    `json:"sales_invoice_id"`
	Tag            string  `json:"tag"` // PAID | ONGOING
	Customer       string  `json:"customer"`
	CustomerCode   string  `json:"customer_code"`
	SalesInvoiceNo string  `json:"sales_invoice_no"`
	DocDate        string  `json:"doc_date"`
	SoNo           string  `json:"so_no"`
	PaymentTerm    string  `json:"payment_term"`
	TotalAmountDue float64 `json:"total_amount_due"`
	Paid           float64 `json:"paid"`
	Balance        float64 `json:"balance"`
	NextDue        string  `json:"next_due"`
}

// One row of the A/P Record of Transactions.
//
// The spec describes the A/R side in full (12.3) but names the A/P one only in
// 3.2's module list for the A/P department, so this mirrors its twin rather than
// inventing a shape. The correspondence is the spec's own: 12.3 states that
// INVOICE DUE on the A/P side is the counterpart of NEXT DUE on the A/R side,
// "likewise typed on the Invoice Receipt and carried forward (5.14)".
//
// Two things follow from that and are deliberate:
//
//   - INVOICE DUE is read straight off the Invoice Receipt and never recomputed.
//     Payment terms are a commercial term, not a date calculator - no setup
//     declares what "30 days" means in days (invariant 18).
//
//   - BALANCE is computed at read time from the receipt less everything applied
//     against it, never stored (12.9). A stored balance drifts the moment any
//     path forgets to update it, and a drifted balance is silently wrong.
type APRecordRow struct {
	InvoiceReceiptId uint    `json:"invoice_receipt_id"`
	Tag              string  `json:"tag"` // PAID | ONGOING
	Supplier         string  `json:"supplier"`
	SupplierCode     string  `json:"supplier_code"`
	InvoiceReceiptNo string  `json:"invoice_receipt_no"`
	ReceiptType      string  `json:"receipt_type"` // INVOICE RECEIPT | BULK INVOICE RECEIPT
	DocDate          string  `json:"doc_date"`
	ReferencePo      string  `json:"reference_po"`
	PaymentTerm      string  `json:"payment_term"`
	NetAmount        float64 `json:"net_amount"`
	Applied          float64 `json:"applied"`
	Balance          float64 `json:"balance"`
	InvoiceDue       string  `json:"invoice_due"`
}

// One row of the Billing list (spec 12.5), keyed on the SO. A row closes when
// BALANCE reaches zero - never on the first Official Receipt, however many a
// partly-paid SO already carries (14.37).
//
// PAID counts what Payment Receipts applied to the SO's invoices (including
// withholding applied); ADJUSTMENTS is the signed sum of the SO's hand-keyed
// ledger (12.4). Both are shown, so a balance can be read rather than guessed.
type BillingRow struct {
	SoNo           string           `json:"so_no"`
	Date           string           `json:"date"`
	ProjectName    string           `json:"project_name"`
	SalesExecutive string           `json:"sales_executive"`
	Customer       string           `json:"customer"`
	TotalDue       float64          `json:"total_due"`
	Adjustments    float64          `json:"adjustments"`
	Paid           float64          `json:"paid"`
	Balance        float64          `json:"balance"`
	NextDue        string           `json:"next_due"`
	Status         string           `json:"status"` // ACTIVE | CLOSED
	Receipts       []BillingReceipt `json:"receipts"`
}

// An Official Receipt recorded against the SO's invoices, with the payment that
// carried it. The OR number is transcribed from the physical receipt (5.22).
type BillingReceipt struct {
	SoNo   string  `json:"so_no"`
	OrNo   string  `json:"or_no"`
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	DocNo  int     `json:"doc_no"`
}
