//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// The Sales quotation's item and model pickers moved server-side (2026-09-16): the item
// picker's "one row per item name" was a catalogue-wide GroupBy in the client, and the model
// picker's "same item name as this one" was a filter over the page's whole ItemList. Both are
// now SQL, which is worth testing directly - the modals cannot show whether the dedup is
// really catalogue-wide or just deduped the page they happened to fetch. Read-only.
//
//	go test -tags dbtest -run TestItemPicker -v ./services/setup_services/
func TestItemPickerNamesIsDedupedAcrossTheWholeCatalogue(t *testing.T) {
	connectForTest(t)

	var distinctNames int64
	if err := initializers.DB.Raw(
		"SELECT COUNT(*) FROM (SELECT MIN(id) AS id FROM vw_items GROUP BY item_name_id) d").
		Scan(&distinctNames).Error; err != nil {
		t.Fatal(err)
	}
	if distinctNames == 0 {
		t.Skip("no items in this database")
	}

	seenName := map[uint]bool{}
	seenItem := map[uint]bool{}
	var collected int64
	page := 1

	for {
		rows, meta, _, err := GetItemPickerNames("", page)
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		if len(rows) == 0 {
			t.Fatalf("page %d came back empty while has_next was still true", page)
		}
		if len(rows) > itemPageSize {
			t.Fatalf("page %d returned %d rows, page size is %d", page, len(rows), itemPageSize)
		}
		if meta.Total != distinctNames {
			t.Errorf("page %d reports total %d, there are %d distinct item names", page, meta.Total, distinctNames)
		}
		if meta.Page != page {
			t.Errorf("asked for page %d, meta says %d", page, meta.Page)
		}
		if (page > 1) != meta.HasPrev {
			t.Errorf("page %d: has_prev %v", page, meta.HasPrev)
		}

		for _, row := range rows {
			// The whole point of the endpoint: one row per item name, across every page.
			if seenName[row.ItemNameId] {
				t.Fatalf("item name %d appears on more than one page (item %d)", row.ItemNameId, row.ID)
			}
			seenName[row.ItemNameId] = true

			if seenItem[row.ID] {
				t.Fatalf("item %d returned twice", row.ID)
			}
			seenItem[row.ID] = true
			collected++
		}

		if !meta.HasNext {
			break
		}
		page++
	}

	if collected != distinctNames {
		t.Fatalf("walked %d rows over %d pages, there are %d distinct item names", collected, page, distinctNames)
	}

	t.Logf("%s: %d distinct item names over %d pages of %d", os.Getenv("DB_NAME"), distinctNames, page, itemPageSize)
}

func TestItemPickerBomIdMatchesTheBomTable(t *testing.T) {
	connectForTest(t)

	rows, _, _, err := GetItemPickerNames("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Skip("no items in this database")
	}

	// bom_id drives BOM vs SINGLE in the modal and which branch the quotation takes, so a
	// wrong value here silently sends a picked item down the wrong path.
	for _, row := range rows {
		var want uint
		if err := initializers.DB.Raw(
			"SELECT ISNULL((SELECT TOP 1 id FROM tbl_setup_item_bom WHERE item_id = ?), 0)", row.ID).
			Scan(&want).Error; err != nil {
			t.Fatal(err)
		}
		if row.BomId != want {
			t.Errorf("item %d: picker says bom_id %d, the BOM table says %d", row.ID, row.BomId, want)
		}
	}
}

func TestItemPickerModelsShareOneItemName(t *testing.T) {
	connectForTest(t)

	names, _, _, err := GetItemPickerNames("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Skip("no items in this database")
	}

	// An item name that actually has several models is the interesting case; fall back to
	// the first row if every name on page 1 has only one.
	subject := names[0]
	for _, candidate := range names {
		var count int64
		if err := initializers.DB.Raw("SELECT COUNT(*) FROM vw_items WHERE item_name_id = ?", candidate.ItemNameId).
			Scan(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count > 1 {
			subject = candidate
			break
		}
	}

	found := false
	page := 1
	var collected int64

	for {
		rows, meta, _, err := GetItemPickerModels(int(subject.ID), "", page)
		if err != nil {
			t.Fatalf("models page %d: %v", page, err)
		}
		if len(rows) == 0 {
			t.Fatalf("item %d returned no models at all - it is at least its own model", subject.ID)
		}

		for _, row := range rows {
			if row.ItemNameId != subject.ItemNameId {
				t.Fatalf("model %d has item_name_id %d, expected %d", row.ID, row.ItemNameId, subject.ItemNameId)
			}
			if row.ID == subject.ID {
				found = true
			}
			collected++
		}

		if !meta.HasNext {
			if meta.Total != collected {
				t.Errorf("walked %d models, meta reports %d", collected, meta.Total)
			}
			break
		}
		page++
	}

	if !found {
		t.Errorf("item %d is missing from its own model list", subject.ID)
	}

	// A bad item id is a client bug, not an empty list.
	if _, _, status, err := GetItemPickerModels(0, "", 1); err == nil || status != 400 {
		t.Errorf("item_id 0 returned status %d, err %v - want a 400", status, err)
	}

	// An id that is simply not in the catalogue is the project-template case: the template
	// keeps the item id it was built from, and a rebuilt database renumbers the items. It
	// has to be reported, because an empty list reads as "this component has no models".
	var missing int64
	if err := initializers.DB.Raw("SELECT MAX(id) + 1000 FROM vw_items").Scan(&missing).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, status, err := GetItemPickerModels(int(missing), "", 1); err == nil || status != 404 {
		t.Errorf("item %d is not in the catalogue but returned status %d, err %v - want a 404",
			missing, status, err)
	}

	t.Logf("%s: item %d (%s) has %d models", os.Getenv("DB_NAME"), subject.ID, subject.ItemName, collected)
}

// The pickers page 20 at a time, so the order decides what a user can actually find. In id
// order the 15 variants of pump model NM 40/12 were spread over pages 42-47 of 123; by model
// they sit together. This walks every page of the biggest model list and checks the order
// holds across page boundaries - and that paging loses or repeats nothing, which is what a
// wrong ORDER BY with OFFSET/FETCH does. Read-only.
func TestItemPickerModelsAreOrderedByModel(t *testing.T) {
	connectForTest(t)

	// The longest model list is where order matters and where a paging fault shows up.
	var subjectID int64
	if err := initializers.DB.Raw(
		`SELECT TOP 1 MIN(id) FROM vw_items WHERE item_name_id <> 0
		 GROUP BY item_name_id ORDER BY COUNT(*) DESC`).Scan(&subjectID).Error; err != nil {
		t.Fatal(err)
	}
	if subjectID == 0 {
		t.Skip("no items in this database")
	}

	// The expected order comes from SQL Server rather than a Go string compare: sorting is the
	// database's collation, which weighs punctuation differently from Go's byte-wise <, so
	// "BNM 32/16B-60/A" and "BNM 32/16B60/A" legitimately order differently in the two.
	var want []uint
	if err := initializers.DB.Raw(
		`SELECT id FROM vw_items WHERE item_name_id = (SELECT item_name_id FROM vw_items WHERE id = ?)
		 ORDER BY item_model ASC, id ASC`, subjectID).Scan(&want).Error; err != nil {
		t.Fatal(err)
	}

	var got []uint
	page := 1
	var total int64

	for {
		rows, pagination, _, err := GetItemPickerModels(int(subjectID), "", page)
		if err != nil {
			t.Fatalf("models page %d: %v", page, err)
		}
		total = pagination.Total

		for _, row := range rows {
			got = append(got, row.ID)
		}

		if !pagination.HasNext {
			break
		}
		page++
	}

	if int64(len(got)) != total || len(got) != len(want) {
		t.Fatalf("walked %d models over %d pages, meta reports %d, the database holds %d",
			len(got), page, total, len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("position %d (page %d) is item %d, model order puts item %d there",
				i, i/itemPageSize+1, got[i], want[i])
		}
	}

	t.Logf("%s: %d models over %d pages, in model order", os.Getenv("DB_NAME"), len(got), page)
}

func TestItemPickerNamesAreOrderedByName(t *testing.T) {
	connectForTest(t)

	// Compared against the database's own ordering, for the collation reason above.
	var want []uint
	if err := initializers.DB.Raw(
		`SELECT id FROM vw_items WHERE id IN (SELECT MIN(id) FROM vw_items GROUP BY item_name_id)
		 ORDER BY item_name ASC, id ASC`).Scan(&want).Error; err != nil {
		t.Fatal(err)
	}
	if len(want) == 0 {
		t.Skip("no items in this database")
	}

	var got []uint
	page := 1

	for {
		rows, pagination, _, err := GetItemPickerNames("", page)
		if err != nil {
			t.Fatalf("names page %d: %v", page, err)
		}

		for _, row := range rows {
			got = append(got, row.ID)
		}

		if !pagination.HasNext {
			break
		}
		page++
	}

	if len(got) != len(want) {
		t.Fatalf("walked %d names, the database holds %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("position %d is item %d, name order puts item %d there", i, got[i], want[i])
		}
	}
}

func TestItemPickerSearchNarrows(t *testing.T) {
	connectForTest(t)

	all, allMeta, _, err := GetItemPickerNames("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Skip("no items in this database")
	}

	term := all[0].ItemCode
	found, meta, _, err := GetItemPickerNames(term, 1)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Total == 0 {
		t.Fatalf("searching for item code %q matched nothing", term)
	}
	if meta.Total > allMeta.Total {
		t.Errorf("search for %q matched %d, more than the unfiltered %d", term, meta.Total, allMeta.Total)
	}
	if len(found) > itemPageSize {
		t.Errorf("search returned %d rows, page size is %d", len(found), itemPageSize)
	}
}
