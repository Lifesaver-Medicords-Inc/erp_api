//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// Removing a tab must take everything filed under it - content, Size Up, Final Selection,
// Advanced Conditions, items, wiring - children before parents, so the foreign keys do not
// refuse it. Runs inside a transaction that is always rolled back. Run by hand:
//
//	go test -tags dbtest -run TestDeleteItemSetChildren -v ./services/sales_services/
func TestDeleteItemSetChildren(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis() // the cascade writes audit rows through DbInsert, which clears cache keys

	// An item set with Size Up or Final rows, so the foreign-key ordering is exercised.
	var found int
	initializers.DB.Raw(`
		SELECT TOP 1 c.based_id
		  FROM tbl_trans_sales_project_content c
		  JOIN tbl_trans_sales_project_item_set s ON s.item_set_id = c.based_id
		 WHERE c.based_id > 0 AND (EXISTS (SELECT 1 FROM tbl_trans_sales_project_size_up u WHERE u.sales_project_content_id = c.content_id)
		    OR EXISTS (SELECT 1 FROM tbl_trans_sales_project_content_final f WHERE f.sales_project_content_id = c.content_id))
		 ORDER BY c.based_id DESC`).Scan(&found)
	itemSetID := uint(found)
	if itemSetID == 0 {
		t.Skip("no item set with Size Up or Final rows on this database")
	}

	tx := initializers.DB.Begin()
	defer tx.Rollback()

	count := func(model interface{}, where string) int64 {
		var n int64
		tx.Model(model).Where(where, itemSetID).Count(&n)
		return n
	}
	const byContent = "sales_project_content_id IN (SELECT content_id FROM tbl_trans_sales_project_content WHERE based_id = ?)"

	t.Logf("item set %d before: content %d, size up %d, final %d, conditions %d, items %d, wiring %d",
		itemSetID,
		count(&models.SalesProjectContent{}, "based_id = ?"),
		count(&models.SalesProjectSizeUp{}, byContent),
		count(&models.SalesProjectContentFinal{}, byContent),
		count(&models.SalesProjectAdvancedConditions{}, "based_id = ?"),
		count(&models.SalesProjectItems{}, "based_id = ?"),
		count(&models.SalesProjectWiring{}, "based_id = ?"))

	if err := deleteItemSetChildren(tx, itemSetID, models.At{}); err != nil {
		t.Fatalf("deleteItemSetChildren: %v", err)
	}
	if err := tx.Where("item_set_id = ?", itemSetID).Delete(&models.SalesProjectItemSet{}).Error; err != nil {
		t.Fatalf("item set delete after its children: %v", err)
	}

	left := map[string]int64{
		"content":    count(&models.SalesProjectContent{}, "based_id = ?"),
		"conditions": count(&models.SalesProjectAdvancedConditions{}, "based_id = ?"),
		"items":      count(&models.SalesProjectItems{}, "based_id = ?"),
		"wiring":     count(&models.SalesProjectWiring{}, "based_id = ?"),
	}
	for what, n := range left {
		if n != 0 {
			t.Errorf("%s rows left under item set %d: %d", what, itemSetID, n)
		}
	}
}
