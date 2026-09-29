//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// What an existing quote's save does to its Size Up / FINAL lists (syncContentChildren,
// the UPDATED path every saved project quote takes). Rolled back.
//
//	go test -tags dbtest -run TestContentChildrenSync -v ./services/sales_services/
func TestContentChildrenSync(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	var contentID uint
	initializers.DB.Raw(`SELECT TOP 1 content_id FROM tbl_trans_sales_project_content ORDER BY content_id DESC`).Scan(&contentID)
	if contentID == 0 {
		t.Skip("no content rows")
	}
	var pumps []struct {
		ID        uint
		ItemModel string
	}
	initializers.DB.Raw(`SELECT TOP 3 id, item_model FROM vw_items WHERE item_name = 'PUMP' ORDER BY id`).Scan(&pumps)
	if len(pumps) < 3 {
		t.Skip("need 3 pumps")
	}

	tx := initializers.DB.Begin()
	defer tx.Rollback()

	count := func(label string) (finals []models.SalesProjectContentFinal, sizeUps []models.SalesProjectSizeUp) {
		tx.Where("sales_project_content_id = ?", contentID).Order("id").Find(&finals)
		tx.Where("sales_project_content_id = ?", contentID).Order("id").Find(&sizeUps)
		t.Logf("%-40s finals=%d size_ups=%d", label, len(finals), len(sizeUps))
		return
	}

	// Start from a known state: no children on this content row.
	tx.Where("sales_project_content_id = ?", contentID).Delete(&models.SalesProjectContentFinal{})
	tx.Where("sales_project_content_id = ?", contentID).Delete(&models.SalesProjectSizeUp{})

	final := func(i int, id uint) models.SalesProjectContentFinal {
		f := models.SalesProjectContentFinal{ID: id}
		f.ItemID, f.Final, f.Fla, f.Voltage = pumps[i].ID, pumps[i].ItemModel, 5, 220
		return f
	}
	sizeUp := func(i int, id uint) models.SalesProjectSizeUp {
		s := models.SalesProjectSizeUp{ID: id}
		s.ItemID, s.Model = pumps[i].ID, pumps[i].ItemModel
		return s
	}

	// 1. A screen picks one pump for Size Up and FINAL on a saved quote (new rows, id 0).
	if err := syncContentChildren(tx, contentID,
		[]models.SalesProjectContentFinal{final(0, 0)}, []models.SalesProjectSizeUp{sizeUp(0, 0)},
		true, true, nil, nil, models.At{}); err != nil {
		t.Fatal(err)
	}
	finals, sizeUps := count("1. pump A picked (new rows)")
	if len(finals) != 1 || len(sizeUps) != 1 {
		t.Fatalf("a newly picked pump must be stored: finals=%d size_ups=%d, want 1/1", len(finals), len(sizeUps))
	}
	idA, sizeUpA := finals[0].ID, sizeUps[0].ID
	ids := func(rows []models.SalesProjectContentFinal) []uint {
		out := []uint{}
		for _, r := range rows {
			out = append(out, r.ID)
		}
		return out
	}

	// 2. The same screen saves again before it has been told A's id: A sent with id 0 again.
	if err := syncContentChildren(tx, contentID,
		[]models.SalesProjectContentFinal{final(0, 0)}, []models.SalesProjectSizeUp{sizeUp(0, 0)},
		true, true, []uint{}, []uint{}, models.At{}); err != nil {
		t.Fatal(err)
	}
	finals, sizeUps = count("2. A re-sent without its id")
	if len(finals) != 1 || finals[0].ID != idA || len(sizeUps) != 1 || sizeUps[0].ID != sizeUpA {
		t.Fatalf("re-sending A without its id must update A, not add a copy: finals=%v", ids(finals))
	}

	// 3. The other app, which has read A, adds B.
	if err := syncContentChildren(tx, contentID,
		[]models.SalesProjectContentFinal{final(0, idA), final(1, 0)}, nil,
		true, false, []uint{idA}, nil, models.At{}); err != nil {
		t.Fatal(err)
	}
	finals, _ = count("3. other app adds B")
	if len(finals) != 2 {
		t.Fatalf("want A and B, got %v", ids(finals))
	}

	// 4. The first screen never saw B, and picks C: it sends A and C, knowing only A.
	if err := syncContentChildren(tx, contentID,
		[]models.SalesProjectContentFinal{final(0, idA), final(2, 0)}, nil,
		true, false, []uint{idA}, nil, models.At{}); err != nil {
		t.Fatal(err)
	}
	finals, _ = count("4. first screen adds C, never saw B")
	if len(finals) != 3 {
		t.Fatalf("B was added by the other app and must survive a save from a screen that never saw it; got %v", ids(finals))
	}

	// 5. A screen that has seen all three takes C off.
	all := ids(finals)
	var keep []models.SalesProjectContentFinal
	for _, f := range finals {
		if f.ItemID != pumps[2].ID {
			keep = append(keep, f)
		}
	}
	if err := syncContentChildren(tx, contentID, keep, nil, true, false, all, nil, models.At{}); err != nil {
		t.Fatal(err)
	}
	finals, _ = count("5. a screen that saw A, B, C removes C")
	if len(finals) != 2 {
		t.Fatalf("a row the screen had seen and took off must be deleted; got %v", ids(finals))
	}

	// 6. An older client sends no known ids: the stored list becomes exactly what it sent.
	if err := syncContentChildren(tx, contentID,
		[]models.SalesProjectContentFinal{final(0, idA)}, nil, true, false, nil, nil, models.At{}); err != nil {
		t.Fatal(err)
	}
	finals, _ = count("6. older client (no known ids) sends only A")
	if len(finals) != 1 || finals[0].ID != idA {
		t.Fatalf("without known ids the old replace rule applies; got %v", ids(finals))
	}
}
