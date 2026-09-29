//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// Walks every BOM page forward and back: 20 heads a page in id order, none twice, all of them
// reached, each page carrying exactly its own heads' component lines, and ?at= opening a page
// on the BOM asked for. Then one search. Reads only. For the full catalogue run it against
// test_fresh (godotenv does not override a variable already set):
//
//	DB_NAME=Lightspeed_test_fresh go test -tags dbtest -run TestBomPages -v ./services/setup_services/
func TestBomPages(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	var total int64
	if err := initializers.DB.Model(&models.BomViewList{}).Count(&total).Error; err != nil {
		t.Fatal(err)
	}
	// Lines of BOMs that exist. test_fresh has one leftover line (id 1) whose BOM was never
	// created; no page can carry it, and the old screen never showed it either.
	var totalLines int64
	if err := initializers.DB.Model(&models.BomViewItemList{}).
		Where("item_bom_id IN (SELECT id FROM vw_get_bom_list)").Count(&totalLines).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("%d BOMs, %d component lines", total, totalLines)

	seen := map[uint]bool{}
	var firstIds []uint
	var lines int64
	after := 0
	pages := 0
	for {
		page, meta, _, err := GetBomsPaged(after, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.ItemBom) == 0 {
			break
		}
		pages++
		if len(page.ItemBom) > bomPageSize {
			t.Fatalf("page %d has %d heads", pages, len(page.ItemBom))
		}

		heads := map[uint]bool{}
		var last uint
		for _, h := range page.ItemBom {
			if seen[h.ID] {
				t.Fatalf("BOM %d came back twice", h.ID)
			}
			if h.ID <= last {
				t.Fatalf("page %d is out of order at %d", pages, h.ID)
			}
			seen[h.ID], heads[h.ID], last = true, true, h.ID
		}
		for _, d := range page.ItemDetails {
			if !heads[d.ItemBomID] {
				t.Fatalf("page %d carries a line of BOM %d, which is not on it", pages, d.ItemBomID)
			}
		}
		lines += int64(len(page.ItemDetails))
		firstIds = append(firstIds, page.ItemBom[0].ID)

		if meta.Page != pages {
			t.Errorf("page %d reports itself as page %d", pages, meta.Page)
		}
		if !meta.HasNext {
			if int(total) != len(seen) {
				t.Fatalf("has_next false after %d of %d BOMs", len(seen), total)
			}
			break
		}
		after = int(last)
	}
	t.Logf("forward: %d pages, %d BOMs, %d lines", pages, len(seen), lines)
	if int64(len(seen)) != total || lines != totalLines {
		t.Errorf("walked %d BOMs / %d lines, want %d / %d", len(seen), lines, total, totalLines)
	}

	// Back again from the last page with ?before=: the same pages, in reverse.
	if pages > 1 {
		before := int(firstIds[len(firstIds)-1])
		for i := len(firstIds) - 2; i >= 0; i-- {
			page, _, _, err := GetBomsPaged(0, before, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(page.ItemBom) == 0 || page.ItemBom[0].ID != firstIds[i] {
				t.Fatalf("walking back, page %d starts at %v, want %d", i+1, page.ItemBom, firstIds[i])
			}
			before = int(page.ItemBom[0].ID)
		}
		t.Logf("backward: the same %d pages", pages)
	}

	// ?at= opens a page on the BOM asked for.
	var middle uint
	for id := range seen {
		middle = id
		break
	}
	if middle > 0 {
		page, _, _, err := GetBomsPaged(0, 0, int(middle))
		if err != nil {
			t.Fatal(err)
		}
		if len(page.ItemBom) == 0 || page.ItemBom[0].ID != middle {
			t.Errorf("?at=%d opened a page starting at %v", middle, page.ItemBom)
		}
	}

	rows, meta, _, err := GetBomsSearch("", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("search, no term: %d on page 1 of %d (%d matches)", len(rows), meta.TotalPages, meta.Total)
	if meta.Total != total || len(rows) > bomPageSize {
		t.Errorf("search found %d of %d, page of %d", meta.Total, total, len(rows))
	}
}
