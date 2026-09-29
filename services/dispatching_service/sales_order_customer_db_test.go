//go:build dbtest

package dispatching_services

import (
	"os"
	"strings"
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// The dispatching Sales Order list names every order's customer (2026-09-29: CUSTOMER was
// blank on all of them - customer_name is never filled in on the order). Reads only.
//
//	go test -tags dbtest -run TestSalesOrdersNameTheirCustomer -v ./services/dispatching_service/
func TestSalesOrdersNameTheirCustomer(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	svc := NewSalesOrderService()
	orders, _, err := svc.GetSalesOrdersService(nil)
	if err != nil {
		t.Fatalf("reading the orders: %v", err)
	}
	if len(orders) == 0 {
		t.Skip("no sales orders in this database")
	}

	for _, o := range orders {
		t.Logf("SO#%s  customer %d  %q", o.Doc, o.Customer_ID, o.CustomerName)
		if o.Customer_ID != 0 && strings.TrimSpace(o.CustomerName) == "" {
			t.Errorf("SO#%s (customer %d) has no customer name", o.Doc, o.Customer_ID)
		}

		one, _, err := svc.GetSalesOrderService(map[string]interface{}{"order_id": o.Order_ID})
		if err != nil {
			t.Errorf("SO#%s on its own: %v", o.Doc, err)
			continue
		}
		if one.CustomerName != o.CustomerName {
			t.Errorf("SO#%s: the list says %q, the single read %q", o.Doc, o.CustomerName, one.CustomerName)
		}
	}
}
