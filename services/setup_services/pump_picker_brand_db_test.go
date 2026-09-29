//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// The Size Up / Final pump picker's BRAND column (user-reported 2026-09-28: blank for every
// pump). Every picker row must carry the same brand vw_items gives that item, and an item with
// no brand must come back as "" rather than failing the scan.
//
// Read-only. Run by hand:
//
//	go test -tags dbtest -run TestPumpPickerBrand -v ./services/setup_services/
func TestPumpPickerBrand(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	rows, _, err := GetPumpPickerRows("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Skip("no pump items on this database")
	}

	var brands []struct {
		ID        uint
		ItemBrand string
	}
	if err := initializers.DB.Raw(
		"SELECT id, ISNULL(item_brand, '') AS item_brand FROM vw_items WHERE item_name = 'PUMP'").
		Scan(&brands).Error; err != nil {
		t.Fatal(err)
	}
	want := map[uint]string{}
	for _, b := range brands {
		want[b.ID] = b.ItemBrand
	}

	withBrand := 0
	for _, row := range rows {
		if row.ItemBrand != want[row.ID] {
			t.Fatalf("item %d (%s): picker brand %q, vw_items brand %q", row.ID, row.ItemModel, row.ItemBrand, want[row.ID])
		}
		if row.ItemBrand != "" {
			withBrand++
		}
	}
	if withBrand == 0 {
		t.Fatalf("none of the %d pump rows carries a brand", len(rows))
	}
	t.Logf("%d pump rows, %d with a brand, every one matching vw_items (e.g. %s -> %q)",
		len(rows), withBrand, rows[0].ItemModel, rows[0].ItemBrand)
}
