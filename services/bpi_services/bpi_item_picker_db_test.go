//go:build dbtest

package bpi_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// BPI's Add Item picker pages server-side as of 2026-09-16 (it used to download every row of
// vw_bpi_item_list and filter in memory). Read-only. Run by hand:
//
//	go test -tags dbtest -run TestBpiItemList -v ./services/bpi_services/
func connectForTest(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
}

func TestBpiItemListPagedWalksEveryRowOnce(t *testing.T) {
	connectForTest(t)

	var total int64
	if err := initializers.DB.Model(&models.BpiItemList{}).Count(&total).Error; err != nil {
		t.Fatal(err)
	}
	if total == 0 {
		t.Skip("vw_bpi_item_list is empty in this database")
	}

	seen := map[uint]bool{}
	var collected int64
	page := 1

	for {
		rows, meta, _, err := GetBpiItemListPaged("", page)
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		if len(rows) == 0 {
			t.Fatalf("page %d came back empty while has_next was still true", page)
		}
		if len(rows) > bpiItemPageSize {
			t.Fatalf("page %d returned %d rows, page size is %d", page, len(rows), bpiItemPageSize)
		}
		if meta.Total != total {
			t.Errorf("page %d reports total %d, the view has %d rows", page, meta.Total, total)
		}
		if meta.Page != page {
			t.Errorf("asked for page %d, meta says %d", page, meta.Page)
		}
		if (page > 1) != meta.HasPrev {
			t.Errorf("page %d: has_prev %v", page, meta.HasPrev)
		}

		for _, row := range rows {
			if seen[row.ID] {
				t.Fatalf("item %d returned on more than one page", row.ID)
			}
			seen[row.ID] = true
			collected++
		}

		if !meta.HasNext {
			break
		}
		page++
	}

	if collected != total {
		t.Fatalf("walked %d rows over %d pages, the view has %d", collected, page, total)
	}

	t.Logf("%s: %d bpi item rows over %d pages of %d", os.Getenv("DB_NAME"), total, page, bpiItemPageSize)
}

// The picker's "Add Selected Items" reads price, descriptions and the two status fields off
// every ticked row, so a narrowed projection here would break that path - which is exactly
// the KeyNotFoundException vw_bpi_item_list was widened to fix.
func TestBpiItemListPagedReturnsTheWholeRow(t *testing.T) {
	connectForTest(t)

	rows, _, _, err := GetBpiItemListPaged("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Skip("vw_bpi_item_list is empty in this database")
	}

	for _, row := range rows {
		var want models.BpiItemList
		if err := initializers.DB.Where("id = ?", row.ID).Take(&want).Error; err != nil {
			t.Fatalf("item %d: %v", row.ID, err)
		}
		if row.ItemCode != want.ItemCode || row.GeneralName != want.GeneralName ||
			row.ItemPrice != want.ItemPrice || row.StatusTangible != want.StatusTangible ||
			row.StatusTrade != want.StatusTrade || row.ShortDesc != want.ShortDesc {
			t.Fatalf("item %d came back with fields that differ from the view row", row.ID)
		}
	}
}

func TestBpiItemListSearchNarrows(t *testing.T) {
	connectForTest(t)

	all, allMeta, _, err := GetBpiItemListPaged("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Skip("vw_bpi_item_list is empty in this database")
	}

	term := all[0].ItemCode
	found, meta, _, err := GetBpiItemListPaged(term, 1)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Total == 0 {
		t.Fatalf("searching for item code %q matched nothing", term)
	}
	if meta.Total > allMeta.Total {
		t.Errorf("search for %q matched %d rows, more than the unfiltered %d", term, meta.Total, allMeta.Total)
	}
	if len(found) > bpiItemPageSize {
		t.Errorf("search returned %d rows, page size is %d", len(found), bpiItemPageSize)
	}

	t.Logf("%s: %q matched %d of %d rows", os.Getenv("DB_NAME"), term, meta.Total, allMeta.Total)
}
