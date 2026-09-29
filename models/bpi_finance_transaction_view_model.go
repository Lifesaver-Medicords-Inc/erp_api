package models

// BpiFinanceTransactionView is one row of the BPI Finance tab's TRANSACTIONS table (spec 4.1.7):
// one of the partner's sales orders with the documents raised against it. The *_dates columns
// hold one line per document, lined up with the document column beside them. TotalAmountDue,
// Paid and Balance are what the SO is worth, what has been paid on its invoices, and the
// difference; the tab's ACCOUNT BALANCE is the sum of Balance (2026-09-29).
type BpiFinanceTransactionView struct {
	OrderID                    uint    `json:"order_id"`
	SalesOrder                 string  `json:"sales_order"`
	SalesOrderDate             string  `json:"sales_order_date"`
	PurchaseOrders             string  `json:"purchase_orders"`
	PurchaseOrderDates         string  `json:"purchase_order_dates"`
	SalesInvoices              string  `json:"sales_invoices"`
	SalesInvoiceDates          string  `json:"sales_invoice_dates"`
	PaymentReceipts            string  `json:"payment_receipts"`
	PaymentReceiptDates        string  `json:"payment_receipt_dates"`
	TotalAmountDue             float64 `json:"total_amount_due"`
	Paid                       float64 `json:"paid"`
	Balance                    float64 `json:"balance"`
	Status                     string  `json:"status"`
	CustomerId                 uint    `json:"customer_id"`
	FinanceTransactionBranchId uint    `json:"finance_transaction_branch_id"`
}

func (BpiFinanceTransactionView) TableName() string {
	return "vw_get_bpi_finance_transactions"
}
