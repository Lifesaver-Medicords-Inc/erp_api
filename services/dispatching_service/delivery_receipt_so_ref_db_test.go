//go:build dbtest

package dispatching_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// A delivery receipt's logistics schedule and route name the sales order by its own number
// (doc), never by the quotation code the order's document_no holds (2026-09-30). Reads only.
//
//	go test -tags dbtest -run TestDeliveryReceiptReferencesTheOrderNumber -v ./services/dispatching_service/
func TestDeliveryReceiptReferencesTheOrderNumber(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	var orders []models.Order
	if err := initializers.DB.Select("order_id", "doc", "document_no").Find(&orders).Error; err != nil {
		t.Fatal(err)
	}
	if len(orders) == 0 {
		t.Skip("no sales orders in this database")
	}

	for _, o := range orders {
		got := getSalesOrderDocNo(initializers.DB, o.Order_ID)
		t.Logf("order %d: doc %q, document_no %q -> reference %q", o.Order_ID, o.Doc, o.Document_No, got)
		if got != o.Doc {
			t.Errorf("order %d is referenced as %q, want its own number %q", o.Order_ID, got, o.Doc)
		}
	}

	if got := getSalesOrderDocNo(initializers.DB, 0); got != "" {
		t.Errorf("no order gives %q, want empty", got)
	}
}
