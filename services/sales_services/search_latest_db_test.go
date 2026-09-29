//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// Project and Quick Quote Search list each document once - its latest copy - however many
// EDIT sub-versions it has (user-reported 2026-09-26: Makati Building listed as 1.0, 1.1 and
// 1.2). Reads only. Run by hand, e.g. against test_fresh:
//
//	DB_NAME=Lightspeed_test_fresh go test -tags dbtest -run TestSearchListsLatestCopyOnly -v ./services/sales_services/
func TestSearchListsLatestCopyOnly(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	check := func(name string, search func(string, int) (quotationListShape, error)) {
		shape, err := search("#", 1)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]string{}
		for _, q := range shape.SalesQuotation {
			code := q.DocumentNo + "." + q.VersionNo + "." + q.SubVersionNo
			if other, ok := seen[q.DocumentNo]; ok {
				t.Errorf("%s: %s is listed twice (%s and %s)", name, q.DocumentNo, other, code)
			}
			seen[q.DocumentNo] = code
		}
		t.Logf("%s: %d documents on page 1", name, len(shape.SalesQuotation))
		for _, code := range seen {
			t.Logf("  %s", code)
		}
	}

	check("projects", func(term string, page int) (quotationListShape, error) {
		shape, _, _, err := SearchProjects(term, page)
		return shape, err
	})
	check("quick quotes", func(term string, page int) (quotationListShape, error) {
		shape, _, _, err := SearchQuotations(term, page)
		return shape, err
	})
}
