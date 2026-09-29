//go:build dbtest

package bpi_services

import (
	"testing"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// The BPI Finance tab's two tables (spec 4.1.7), as the API reads them: PENDING lists each
// ongoing quote once, priced at TOTAL AMOUNT DUE, and no quote whose SO is already invoiced;
// TRANSACTIONS lists the finished SOs with their documents. Reads only.
//
//	go test -tags dbtest -run TestBpiFinanceViews -v ./services/bpi_services/
func TestBpiFinanceViews(t *testing.T) {
	connectForTest(t)

	var pending []models.BpiFinancePendingView
	if err := services.DbGetNoCache(&pending, nil); err != nil {
		t.Fatalf("reading PENDING: %v", err)
	}
	seen := map[string]bool{}
	var total float64
	for _, p := range pending {
		if seen[p.QouteRef] {
			t.Errorf("%s is listed more than once - it would be counted twice in ACCOUNT BALANCE", p.QouteRef)
		}
		seen[p.QouteRef] = true
		total += p.TotalPrice
	}
	t.Logf("PENDING: %d quotes, total %.2f", len(pending), total)

	var transactions []models.BpiFinanceTransactionView
	if err := services.DbGetNoCache(&transactions, nil); err != nil {
		t.Fatalf("reading TRANSACTIONS: %v", err)
	}
	t.Logf("TRANSACTIONS: %d finished sales orders", len(transactions))

	finished := map[string]bool{}
	for _, tr := range transactions {
		t.Logf("  %s  %s  SI %q  PAYR %q", tr.SalesOrder, tr.SalesOrderDate, tr.SalesInvoices, tr.PaymentReceipts)
		if tr.SalesOrder == "" || tr.FinanceTransactionBranchId == 0 {
			t.Errorf("a transaction row came back without its SO or branch: %+v", tr)
		}
		finished[tr.SalesOrder] = true
	}
	for _, p := range pending {
		if p.SoRef != "" && finished[p.SoRef] {
			t.Errorf("%s (%s) is finished but still under PENDING, counted in the balance", p.QouteRef, p.SoRef)
		}
	}
}
