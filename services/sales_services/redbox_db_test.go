//go:build dbtest

package sales_services

import (
	"os"
	"strings"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// The Sales red box's quote cards (spec 9): one card per quote - never one per copy - for
// Quick and Project quotes alike, gone once an SO is made from any copy. Reads only. Run by hand:
//
//	go test -tags dbtest -run TestSalesRedBoxQuotations -v ./services/sales_services/
func TestSalesRedBoxQuotations(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	cards, _, err := GetSalesRedBoxQuotations()
	if err != nil {
		t.Fatal(err)
	}

	bare := func(no string) string {
		return strings.NewReplacer("FQ#", "", "Q#", "").Replace(no)
	}

	seen := map[string]bool{}
	projects := 0
	for _, c := range cards {
		if seen[bare(c.DocumentNo)] {
			t.Errorf("%s has more than one card", bare(c.DocumentNo))
		}
		seen[bare(c.DocumentNo)] = true
		if c.ProjectName != "" {
			projects++
		}
	}
	t.Logf("%d cards (%d project quotes)", len(cards), projects)

	// Every quote with no SO is on the rail (bar Abandoned/Lost), and none with one.
	var quotes []models.SalesQuotation
	if err := initializers.DB.Find(&quotes).Error; err != nil {
		t.Fatal(err)
	}
	var orders []struct{ QuotationID uint }
	if err := initializers.DB.Table("tbl_trans_sales_order").Select("quotation_id").Find(&orders).Error; err != nil {
		t.Fatal(err)
	}
	byID := map[uint]string{}
	for _, q := range quotes {
		byID[q.ID] = bare(q.DocumentNo)
	}
	withSO := map[string]bool{}
	for _, o := range orders {
		if no, ok := byID[o.QuotationID]; ok {
			withSO[no] = true
		}
	}
	for no := range withSO {
		if seen[no] {
			t.Errorf("%s already has an SO but still has a quote card", no)
		}
	}
	var closed []struct{ DocumentNo string }
	initializers.DB.Table("tbl_trans_sales_opportunity").Select("document_no").
		Where("status IN ?", []string{"Abandoned", "Lost"}).Find(&closed)
	lost := map[string]bool{}
	for _, c := range closed {
		lost[bare(c.DocumentNo)] = true
	}
	for _, q := range quotes {
		no := bare(q.DocumentNo)
		if !seen[no] && !withSO[no] && !lost[no] {
			t.Errorf("%s has no SO and is not lost, but has no card", no)
		}
	}
}
