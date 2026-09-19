//go:build dbtest

package services_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/bpi_services"
	"github.com/pierceperado/smpc/services/sales_services"
	"github.com/pierceperado/smpc/services/setup_services"
)

// What the sales quotation page loads when it opens, timed on the server: each call cold
// (this database's cache emptied first) and warm (served from Redis), with the size of the
// JSON the client then downloads and parses. Clears ONLY the target database's cache
// namespace. Point it at a database with DB_NAME:
//
//	go test -tags dbtest -run TestQuotationPageLoadTiming -v ./services/
func TestQuotationPageLoadTiming(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir(".."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	calls := []struct {
		name string
		get  func() (interface{}, int, error)
	}{
		{"GET /setup/item (catalogue)", func() (interface{}, int, error) { return setup_services.GetItems(nil) }},
		{"GET /setup/bom", func() (interface{}, int, error) { return setup_services.GetSetupItemBoms(nil) }},
		{"GET /bpi", func() (interface{}, int, error) { return bpi_services.GetBpis(nil) }},
		{"GET /sales/quotation", func() (interface{}, int, error) { return sales_services.GetSalesQuotations(nil) }},
		{"GET /sales/projects", func() (interface{}, int, error) { return sales_services.GetSalesProjects(nil) }},
	}

	if err := services.InvalidateCacheByPattern("model:*"); err != nil {
		t.Fatalf("clearing cache: %v", err)
	}

	var totalCold, totalWarm time.Duration
	var totalBytes int
	for _, call := range calls {
		start := time.Now()
		data, _, err := call.get()
		cold := time.Since(start)
		if err != nil {
			t.Fatalf("%s failed: %v", call.name, err)
		}

		start = time.Now()
		if _, _, err := call.get(); err != nil {
			t.Fatalf("%s (warm) failed: %v", call.name, err)
		}
		warm := time.Since(start)

		body, _ := json.Marshal(data)
		totalCold += cold
		totalWarm += warm
		totalBytes += len(body)
		t.Logf("%-30s cold %7.0f ms   warm %6.0f ms   %8.2f MB", call.name,
			float64(cold.Milliseconds()), float64(warm.Milliseconds()), float64(len(body))/1048576)
	}
	t.Logf("%-30s cold %7.0f ms   warm %6.0f ms   %8.2f MB", "TOTAL",
		float64(totalCold.Milliseconds()), float64(totalWarm.Milliseconds()), float64(totalBytes)/1048576)
}
