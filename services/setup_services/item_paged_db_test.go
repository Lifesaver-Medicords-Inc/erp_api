//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// Item Entry pages through the catalogue 20 at a time instead of downloading all of it
// (2026-09-16). Paging is only useful if walking it start to finish shows every item exactly
// once, and if PREV retraces what NEXT walked - neither of which the UI can prove on its own.
// Read-only. Run by hand:
//
//	go test -tags dbtest -run TestGetItemsPaged -v ./services/setup_services/
//	go test -tags dbtest -run TestGetItemsSearch -v ./services/setup_services/
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

func TestGetItemsPagedWalksTheWholeCatalogue(t *testing.T) {
	connectForTest(t)

	var total int64
	if err := initializers.DB.Model(&models.Item{}).Count(&total).Error; err != nil {
		t.Fatal(err)
	}
	if total == 0 {
		t.Skip("no items in this database")
	}

	// --- forward, the way NEXT >> walks it ---
	seen := make(map[uint]bool, total)
	var forward []uint
	var pages int
	after := 0

	for {
		page, meta, _, err := GetItemsPaged(after, 0, 0)
		if err != nil {
			t.Fatalf("page after id %d: %v", after, err)
		}
		if len(page.Items) == 0 {
			t.Fatalf("page after id %d came back empty while has_next was still true", after)
		}
		if len(page.Items) > itemPageSize {
			t.Fatalf("page after id %d returned %d items, page size is %d", after, len(page.Items), itemPageSize)
		}
		if meta.Total != total {
			t.Errorf("page reports total %d, the table has %d", meta.Total, total)
		}

		pages++
		var last uint
		for i, item := range page.Items {
			if seen[item.ID] {
				t.Fatalf("item %d returned on more than one page", item.ID)
			}
			if i > 0 && item.ID <= last {
				t.Fatalf("page after id %d is not ascending: %d follows %d", after, item.ID, last)
			}
			seen[item.ID] = true
			forward = append(forward, item.ID)
			last = item.ID
		}

		// Every page but the first has something behind it, and only the last has nothing ahead.
		if wantPrev := after > 0; meta.HasPrev != wantPrev {
			t.Errorf("page after id %d: has_prev %v, want %v", after, meta.HasPrev, wantPrev)
		}

		// Children belong to this page's items and to no others - the whole point of the change.
		ids := make(map[uint]bool, len(page.Items))
		for _, item := range page.Items {
			ids[item.ID] = true
		}
		for _, spec := range page.ItemSpecs {
			if !ids[spec.BasedId] {
				t.Fatalf("item specs row %d belongs to item %d, which is not on this page", spec.ID, spec.BasedId)
			}
		}
		for _, production := range page.ItemProductions {
			if !ids[uint(production.ItemId)] {
				t.Fatalf("production row for item %d came back on a page without it", production.ItemId)
			}
		}

		if !meta.HasNext {
			break
		}
		after = int(last)
	}

	if int64(len(forward)) != total {
		t.Fatalf("walked %d items over %d pages, the table has %d", len(forward), pages, total)
	}

	// --- backward, the way << PREV retraces it ---
	var backward []uint
	before := 0

	for {
		page, meta, _, err := GetItemsPaged(0, before, 0)
		if err != nil {
			t.Fatalf("page before id %d: %v", before, err)
		}
		if before == 0 {
			// Seed the walk from the end: the first call with no cursor is the FIRST page, so
			// start from the last id instead and work back from there.
			last := forward[len(forward)-1]
			backward = append(backward, last)
			before = int(last)
			continue
		}
		if len(page.Items) == 0 {
			break
		}

		for i := len(page.Items) - 1; i >= 0; i-- {
			backward = append(backward, page.Items[i].ID)
		}

		// A PREV page must come back in ascending order like any other, despite being read
		// back to front.
		for i := 1; i < len(page.Items); i++ {
			if page.Items[i].ID <= page.Items[i-1].ID {
				t.Fatalf("page before id %d is not ascending: %d follows %d", before, page.Items[i].ID, page.Items[i-1].ID)
			}
		}

		if !meta.HasPrev {
			break
		}
		before = int(page.Items[0].ID)
	}

	if len(backward) != len(forward) {
		t.Fatalf("walked back over %d items, walked forward over %d", len(backward), len(forward))
	}
	for i := range forward {
		if forward[i] != backward[len(backward)-1-i] {
			t.Fatalf("backward walk diverges at %d: %d, forward has %d", i, backward[len(backward)-1-i], forward[i])
		}
	}

	// --- at: the item asked for leads its page, wherever it sits ---
	middle := forward[len(forward)/2]
	page, meta, _, err := GetItemsPaged(0, 0, int(middle))
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) == 0 || page.Items[0].ID != middle {
		t.Fatalf("at=%d opened a page starting at %v", middle, page.Items)
	}
	if meta.Page < 1 || meta.Page > meta.TotalPages {
		t.Errorf("at=%d reports page %d of %d", middle, meta.Page, meta.TotalPages)
	}

	t.Logf("%s: %d items over %d pages of %d, forward and back",
		os.Getenv("DB_NAME"), total, pages, itemPageSize)
}

func TestGetItemsSearchPagesAndFinds(t *testing.T) {
	connectForTest(t)

	// Take, not First: First adds its own "ORDER BY id" on top of the one already given, and
	// SQL Server rejects a column named twice in one ORDER BY.
	var sample models.ItemView
	if err := initializers.DB.Where("item_model <> ''").Order("id ASC").Take(&sample).Error; err != nil {
		t.Skipf("no item with a model to search for: %v", err)
	}

	items, meta, _, err := GetItemsSearch(sample.ItemModel, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatalf("searching for %q returned nothing, though item %d has that model", sample.ItemModel, sample.ID)
	}
	if len(items) > itemPageSize {
		t.Errorf("search returned %d rows, page size is %d", len(items), itemPageSize)
	}
	if meta.Page != 1 || meta.HasPrev {
		t.Errorf("first page reports page %d, has_prev %v", meta.Page, meta.HasPrev)
	}
	if meta.TotalPages < 1 {
		t.Errorf("search reports %d pages for %d matches", meta.TotalPages, meta.Total)
	}

	found := false
	for _, item := range items {
		if item.ID == sample.ID {
			found = true
		}
	}
	if !found && !meta.HasNext {
		t.Errorf("item %d (%s) missing from the only page of its own model search", sample.ID, sample.ItemModel)
	}

	// An empty term is the modal's opening state: every item, still 20 at a time.
	all, allMeta, _, err := GetItemsSearch("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) > itemPageSize {
		t.Errorf("unfiltered search returned %d rows, page size is %d", len(all), itemPageSize)
	}

	// Page 2 must not repeat page 1.
	if allMeta.HasNext {
		second, secondMeta, _, err := GetItemsSearch("", 2)
		if err != nil {
			t.Fatal(err)
		}
		if !secondMeta.HasPrev || secondMeta.Page != 2 {
			t.Errorf("second page reports page %d, has_prev %v", secondMeta.Page, secondMeta.HasPrev)
		}
		first := make(map[uint]bool, len(all))
		for _, item := range all {
			first[item.ID] = true
		}
		for _, item := range second {
			if first[item.ID] {
				t.Fatalf("item %d appears on both search pages", item.ID)
			}
		}
	}

	t.Logf("%s: %q matched %d items over %d pages", os.Getenv("DB_NAME"), sample.ItemModel, meta.Total, meta.TotalPages)
}
