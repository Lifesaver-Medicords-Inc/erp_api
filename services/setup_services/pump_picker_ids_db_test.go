//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// FINAL's pump picker may only offer what SIZE UP lists (spec §5.1.4). It used to pull
// every pump and filter on the client; it now names the ids it wants. This checks the
// filter returns exactly those pumps, in the same order the unfiltered call would, and
// that no ids still means every pump - the SIZE UP picker's own behaviour, which must
// not change.
//
// Read-only. Run by hand:
//
//	go test -tags dbtest -run TestPumpPickerIDs -v ./services/setup_services/
func TestPumpPickerIDs(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	all, _, err := GetPumpPickerRows("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 3 {
		t.Skipf("only %d pump items on this database - need at least 3", len(all))
	}
	t.Logf("no ids: %d pump rows (this is what FINAL used to download to show a handful)", len(all))

	// Three taken from across the list, the way a SIZE UP grid holds a few unrelated pumps.
	want := []int{int(all[0].ID), int(all[len(all)/2].ID), int(all[len(all)-1].ID)}

	got, _, err := GetPumpPickerRows("", want)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ids=%v: %d rows", want, len(got))

	if len(got) != len(want) {
		t.Fatalf("asked for %d ids, got %d rows", len(want), len(got))
	}

	// Same order the unfiltered query gives, since both order by model.
	inWant := map[int]bool{}
	for _, id := range want {
		inWant[id] = true
	}
	expected := []uint{}
	for _, row := range all {
		if inWant[int(row.ID)] {
			expected = append(expected, row.ID)
		}
	}
	for i, row := range got {
		if row.ID != expected[i] {
			t.Fatalf("row %d: got id %d, want %d - the filter changed the order", i, row.ID, expected[i])
		}
		if row.ItemModel == "" {
			t.Fatalf("row %d (id %d) came back with no model", i, row.ID)
		}
	}

	// An id that is not a pump, or not an item at all, must simply not come back - it
	// must never widen the result.
	var notAPump int64
	if err := initializers.DB.Raw(
		"SELECT TOP 1 id FROM vw_items WHERE item_name <> 'PUMP'").Scan(&notAPump).Error; err != nil {
		t.Fatal(err)
	}
	if notAPump > 0 {
		mixed, _, err := GetPumpPickerRows("", []int{want[0], int(notAPump), 999999999})
		if err != nil {
			t.Fatal(err)
		}
		if len(mixed) != 1 || int(mixed[0].ID) != want[0] {
			t.Fatalf("a non-pump id and an unknown id should be ignored, got %d rows", len(mixed))
		}
		t.Logf("non-pump id %d and unknown id 999999999 ignored, as they must be", notAPump)
	}

	// ids and search still narrow together.
	narrowed, _, err := GetPumpPickerRows(all[0].ItemModel, want)
	if err != nil {
		t.Fatal(err)
	}
	if len(narrowed) < 1 {
		t.Fatalf("ids + search returned nothing; expected at least the row searched for")
	}
	for _, row := range narrowed {
		if !inWant[int(row.ID)] {
			t.Fatalf("search widened past the ids: id %d", row.ID)
		}
	}
	t.Logf("ids + search: %d row(s), all within the ids", len(narrowed))
}
