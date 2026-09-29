//go:build dbtest

package bpi_services

import (
	"math"
	"strings"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// The BPI Finance tab's two tables (spec 4.1.7), as the API reads them (rules of 2026-09-29):
// PENDING lists each quote that has no SO yet, once, priced at TOTAL AMOUNT DUE; TRANSACTIONS
// lists every SO, once, with what it is worth, what was paid on its invoices and the balance -
// the figures ACCOUNT BALANCE adds up. Paid is checked against the same sum Billing's A/R makes,
// worked out here on its own. Reads only.
//
//	go test -tags dbtest -run TestBpiFinanceViews -v ./services/bpi_services/
func TestBpiFinanceViews(t *testing.T) {
	connectForTest(t)

	// ---- PENDING ----
	var pending []models.BpiFinancePendingView
	if err := services.DbGetNoCache(&pending, nil); err != nil {
		t.Fatalf("reading PENDING: %v", err)
	}

	// The quote numbers an SO was made from, bare (no Q#/FQ#, no version).
	var converted []string
	if err := initializers.DB.Raw(`
SELECT DISTINCT REPLACE(REPLACE(q.document_no, 'FQ#', ''), 'Q#', '')
  FROM tbl_trans_sales_order s
  JOIN tbl_trans_sales_quotation q ON q.id = s.quotation_id`).Scan(&converted).Error; err != nil {
		t.Fatal(err)
	}
	hasSO := map[string]bool{}
	for _, c := range converted {
		hasSO[c] = true
	}

	bare := func(ref string) string {
		ref = strings.TrimPrefix(strings.TrimPrefix(ref, "FQ#"), "Q#")
		if i := strings.Index(ref, "."); i >= 0 {
			ref = ref[:i]
		}
		return ref
	}

	seen := map[string]bool{}
	for _, p := range pending {
		b := bare(p.QouteRef)
		if seen[b] {
			t.Errorf("%s is listed more than once", p.QouteRef)
		}
		seen[b] = true
		if hasSO[b] {
			t.Errorf("%s already has a sales order but is still under PENDING", p.QouteRef)
		}
	}
	t.Logf("PENDING: %d quotes without a sales order", len(pending))

	// ---- TRANSACTIONS ----
	var transactions []models.BpiFinanceTransactionView
	if err := services.DbGetNoCache(&transactions, nil); err != nil {
		t.Fatalf("reading TRANSACTIONS: %v", err)
	}

	var orderCount int64
	if err := initializers.DB.Table("tbl_trans_sales_order").Count(&orderCount).Error; err != nil {
		t.Fatal(err)
	}
	if int64(len(transactions)) != orderCount {
		t.Errorf("TRANSACTIONS has %d rows for %d sales orders - it must list every SO once", len(transactions), orderCount)
	}

	// What each SO is worth, and what was paid on its invoices, worked out independently: an
	// invoice's paid is Billing's sum (amount + withholding applied), and an invoice belongs to
	// the SO its reference names.
	type orderFigures struct {
		OrderID        uint
		Status         string
		TotalAmountDue float64
		Invoiced       float64
		Paid           float64
	}
	var figures []orderFigures
	if err := initializers.DB.Raw(`
SELECT so.order_id,
       ISNULL(so.status, '') AS status,
       ISNULL(so.total_amount_due, 0) AS total_amount_due,
       ISNULL((SELECT SUM(ISNULL(i.total_amount_due, 0)) FROM tbl_accounting_sales_invoice i
                WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc), 0) AS invoiced,
       ISNULL((SELECT SUM(ISNULL(d.amount_applied, 0) + ISNULL(d.twas_applied, 0))
                 FROM tbl_accounting_payment_receipt_details d
                 JOIN tbl_accounting_sales_invoice i ON i.id = d.sales_invoice_id
                WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc), 0) AS paid
  FROM tbl_trans_sales_order so`).Scan(&figures).Error; err != nil {
		t.Fatal(err)
	}
	want := map[uint]orderFigures{}
	for _, f := range figures {
		want[f.OrderID] = f
	}

	near := func(a, b float64) bool { return math.Abs(a-b) < 0.006 }
	balances := map[uint]float64{}
	listed := map[uint]bool{}
	for _, tr := range transactions {
		t.Logf("  %s  %-9s  due %12.2f  paid %12.2f  balance %12.2f  SI %q  PAYR %q", tr.SalesOrder, tr.Status,
			tr.TotalAmountDue, tr.Paid, tr.Balance, tr.SalesInvoices, tr.PaymentReceipts)
		if tr.SalesOrder == "" || tr.FinanceTransactionBranchId == 0 {
			t.Errorf("a transaction row came back without its SO or branch: %+v", tr)
		}
		if listed[tr.OrderID] {
			t.Errorf("%s is listed more than once", tr.SalesOrder)
		}
		listed[tr.OrderID] = true

		f, ok := want[tr.OrderID]
		if !ok {
			t.Errorf("%s is not a sales order", tr.SalesOrder)
			continue
		}
		due := f.TotalAmountDue
		if strings.EqualFold(f.Status, "CANCELLED") {
			due = f.Invoiced
		}
		if !near(tr.TotalAmountDue, due) {
			t.Errorf("%s: TOTAL AMOUNT DUE %.2f, want %.2f", tr.SalesOrder, tr.TotalAmountDue, due)
		}
		if !near(tr.Paid, f.Paid) {
			t.Errorf("%s: PAID %.2f, want %.2f (Billing's sum over its invoices)", tr.SalesOrder, tr.Paid, f.Paid)
		}
		if !near(tr.Balance, tr.TotalAmountDue-tr.Paid) {
			t.Errorf("%s: BALANCE %.2f is not TOTAL AMOUNT DUE - PAID (%.2f)", tr.SalesOrder, tr.Balance, tr.TotalAmountDue-tr.Paid)
		}
		balances[tr.CustomerId] += tr.Balance
	}
	for customer, balance := range balances {
		t.Logf("ACCOUNT BALANCE of partner %d: %.2f", customer, balance)
	}
}
