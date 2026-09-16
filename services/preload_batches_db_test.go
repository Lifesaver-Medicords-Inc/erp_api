//go:build dbtest

package services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// A preload over more parent rows than SQL Server allows parameters (2,100) must still load, with
// every child exactly once and on the right parent. Read-only. Needs a database with more than
// 2,100 item specs rows - the 2026-09-15 Calpeda specs load gave both Lightspeed_test_fresh and
// Lightspeed_ERP_deploy_rehearsal about 2,450. Run by hand:
//
//	go test -tags dbtest -run TestPreloadInBatches -v ./services/
func TestPreloadInBatches(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir(".."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	var specCount, lineCount int64
	if err := initializers.DB.Model(&models.ItemSpecs{}).Count(&specCount).Error; err != nil {
		t.Fatal(err)
	}
	if err := initializers.DB.Model(&models.ItemSpecsTemplate{}).
		Where("based_id IN (SELECT id FROM tbl_setup_item_specs)").Count(&lineCount).Error; err != nil {
		t.Fatal(err)
	}
	if specCount <= 2100 {
		t.Skipf("only %d item specs rows - not enough to cross the 2,100-parameter limit", specCount)
	}

	var specs []models.ItemSpecs
	if err := fetchRelDB(&specs, map[string]interface{}{}, []string{"ItemSpecsTemplate"}); err != nil {
		t.Fatalf("list preload failed: %v", err)
	}
	if int64(len(specs)) != specCount {
		t.Errorf("loaded %d item specs rows, the table has %d", len(specs), specCount)
	}

	seen := make(map[uint]bool, len(specs))
	var lines int64
	for _, s := range specs {
		if seen[s.ID] {
			t.Fatalf("item specs row %d loaded twice", s.ID)
		}
		seen[s.ID] = true
		for _, line := range s.ItemSpecsTemplate {
			if line.BasedId != s.ID {
				t.Fatalf("spec line %d (based_id %d) attached to item specs row %d", line.ID, line.BasedId, s.ID)
			}
			lines++
		}
	}
	if lines != lineCount {
		t.Errorf("loaded %d spec lines, the table has %d", lines, lineCount)
	}

	// A single record still loads through the one-query path.
	last := specs[len(specs)-1]
	var one models.ItemSpecs
	if err := fetchRelDB(&one, map[string]interface{}{"id": last.ID}, []string{"ItemSpecsTemplate"}); err != nil {
		t.Fatalf("single preload failed: %v", err)
	}
	if one.ID != last.ID || len(one.ItemSpecsTemplate) != len(last.ItemSpecsTemplate) {
		t.Errorf("single record: got id %d with %d lines, want id %d with %d", one.ID, len(one.ItemSpecsTemplate), last.ID, len(last.ItemSpecsTemplate))
	}

	t.Logf("%s: %d item specs rows and %d spec lines loaded in batches of %d",
		os.Getenv("DB_NAME"), specCount, lines, preloadBatchSize)
}
