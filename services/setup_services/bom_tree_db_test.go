//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// GET /setup/bom/tree hands back exactly what walking the full BOM tables would: the BOMs asked
// for, all their lines, and every sub-BOM (a component whose item has a BOM) to any depth -
// no more, no less. Checked for every BOM that nests another, a spread of plain ones, and one
// request of 500 ids at once. Reads only. For the full catalogue run it against test_fresh:
//
//	DB_NAME=Lightspeed_test_fresh go test -tags dbtest -run TestBomTree -v ./services/setup_services/
func TestBomTree(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	var heads []models.BomViewList
	if err := initializers.DB.Find(&heads).Error; err != nil {
		t.Fatal(err)
	}
	var lines []models.BomViewItemList
	if err := initializers.DB.Find(&lines).Error; err != nil {
		t.Fatal(err)
	}

	bomOfItem := map[uint][]uint{}
	for _, h := range heads {
		bomOfItem[h.ItemID] = append(bomOfItem[h.ItemID], h.ID)
	}
	linesOf := map[uint][]models.BomViewItemList{}
	for _, l := range lines {
		linesOf[l.ItemBomID] = append(linesOf[l.ItemBomID], l)
	}

	// The expected tree, walked in memory from the full tables.
	expect := func(ids []int) (map[uint]bool, int) {
		boms := map[uint]bool{}
		queue := []uint{}
		for _, id := range ids {
			queue = append(queue, uint(id))
		}
		lineCount := 0
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			if boms[id] {
				continue
			}
			boms[id] = true
			for _, l := range linesOf[id] {
				lineCount++
				queue = append(queue, bomOfItem[l.ItemID]...)
			}
		}
		return boms, lineCount
	}

	check := func(label string, ids []int) {
		wantBoms, wantLines := expect(ids)
		got, _, err := GetBomTree(ids)
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		gotBoms := map[uint]bool{}
		for _, h := range got.ItemBom {
			if gotBoms[h.ID] {
				t.Errorf("%s: BOM %d came back twice", label, h.ID)
			}
			gotBoms[h.ID] = true
			if !wantBoms[h.ID] {
				t.Errorf("%s: BOM %d is not in the tree", label, h.ID)
			}
		}
		for id := range wantBoms {
			if !gotBoms[id] {
				t.Errorf("%s: BOM %d is missing", label, id)
			}
		}
		for _, l := range got.ItemDetails {
			if !gotBoms[l.ItemBomID] {
				t.Errorf("%s: line %d belongs to BOM %d, which is not in the answer", label, l.Id, l.ItemBomID)
			}
		}
		if len(got.ItemDetails) != wantLines {
			t.Errorf("%s: %d lines, want %d", label, len(got.ItemDetails), wantLines)
		}
	}

	nested := 0
	for _, h := range heads {
		for _, l := range linesOf[h.ID] {
			if len(bomOfItem[l.ItemID]) > 0 {
				nested++
				check("nesting BOM", []int{int(h.ID)})
				break
			}
		}
	}

	step := len(heads)/200 + 1
	plain := 0
	for i := 0; i < len(heads); i += step {
		check("single BOM", []int{int(heads[i].ID)})
		plain++
	}

	var many []int
	for i := 0; i < len(heads) && len(many) < BomTreeMaxIDs; i++ {
		many = append(many, int(heads[i].ID))
	}
	check("500 at once", many)

	t.Logf("%d BOMs, %d lines; checked %d nesting BOMs, %d others, and %d at once", len(heads), len(lines), nested, plain, len(many))
}
