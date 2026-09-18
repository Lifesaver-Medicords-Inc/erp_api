//go:build dbtest

package bpi_services

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// GET /api/bpi is what the sales quotation's customer picker loads (QuotationService
// url_customer "/bpi/"). If any of its nine lists fails, the API answers with an error,
// the quotation keeps an empty customer table, and choosing a customer then throws
// "Cannot find column [customer_code]". Read-only. Point it at a database with the
// usual env vars (they win over .env) and, optionally, BPI_JSON_OUT to keep the body:
//
//	go test -tags dbtest -run TestGetBpisLoads -v ./services/bpi_services/
func TestGetBpisLoads(t *testing.T) {
	connectForTest(t)
	// GetBpis reads through the API's Redis cache, as the API does.
	initializers.InitRedis()

	data, _, err := GetBpis(nil)
	if err != nil {
		t.Fatalf("GetBpis failed: %v", err)
	}

	body, err := json.Marshal(map[string]interface{}{"success": true, "data": data})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	t.Logf("GET /api/bpi body: %d bytes", len(body))

	if out := os.Getenv("BPI_JSON_OUT"); out != "" {
		if err := os.WriteFile(out, body, 0o644); err != nil {
			t.Fatalf("writing %s: %v", out, err)
		}
	}
}
