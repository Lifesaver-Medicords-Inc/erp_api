//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// GetProjectDetail must return each content row WITH its Size Up and Final Selection rows.
// Every project the Sales app opens, and every live refresh, reads through it; a content row
// that arrives without them shows empty grids, and the next save of that content sends empty
// lists - which syncContentChildren reads as "delete them all". Read-only. Run by hand:
//
//	go test -tags dbtest -run TestProjectDetailCarriesContentChildren -v ./services/sales_services/
func TestProjectDetailCarriesContentChildren(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	// A project that has at least one Size Up or Final row, if the database has any.
	var projectID int
	initializers.DB.Raw(`
		SELECT TOP 1 s.based_id
		  FROM tbl_trans_sales_project_item_set s
		  JOIN tbl_trans_sales_project_content c ON c.based_id = s.item_set_id
		 WHERE EXISTS (SELECT 1 FROM tbl_trans_sales_project_size_up u WHERE u.sales_project_content_id = c.content_id)
		    OR EXISTS (SELECT 1 FROM tbl_trans_sales_project_content_final f WHERE f.sales_project_content_id = c.content_id)
		 ORDER BY s.based_id DESC`).Scan(&projectID)
	if projectID == 0 {
		t.Skip("no project on this database has Size Up or Final rows")
	}

	var wantSizeUps, wantFinals int64
	initializers.DB.Model(&models.SalesProjectSizeUp{}).
		Where("sales_project_content_id IN (SELECT c.content_id FROM tbl_trans_sales_project_content c JOIN tbl_trans_sales_project_item_set s ON s.item_set_id = c.based_id WHERE s.based_id = ?)", projectID).
		Count(&wantSizeUps)
	initializers.DB.Model(&models.SalesProjectContentFinal{}).
		Where("sales_project_content_id IN (SELECT c.content_id FROM tbl_trans_sales_project_content c JOIN tbl_trans_sales_project_item_set s ON s.item_set_id = c.based_id WHERE s.based_id = ?)", projectID).
		Count(&wantFinals)

	detail, _, err := GetProjectDetail(projectID)
	if err != nil {
		t.Fatalf("GetProjectDetail(%d): %v", projectID, err)
	}

	var gotSizeUps, gotFinals int
	for _, c := range detail.SalesProjectContent {
		gotSizeUps += len(c.SalesProjectSizeUp)
		gotFinals += len(c.SalesProjectContentFinal)
	}

	t.Logf("project %d: size ups %d in db / %d returned, finals %d in db / %d returned",
		projectID, wantSizeUps, gotSizeUps, wantFinals, gotFinals)

	if int64(gotSizeUps) != wantSizeUps {
		t.Errorf("Size Up rows: database has %d, GetProjectDetail returned %d", wantSizeUps, gotSizeUps)
	}
	if int64(gotFinals) != wantFinals {
		t.Errorf("Final rows: database has %d, GetProjectDetail returned %d", wantFinals, gotFinals)
	}
}
