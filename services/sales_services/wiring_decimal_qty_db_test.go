//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// Wiring QTY and TOTAL QTY keep their decimals through the API (user-reported 2026-09-28: a
// 31.5 m run saved as 0). Runs the widening script's ALTERs, then a new wiring row and a
// field-level update of an existing one, the two ways a quotation save writes wiring - all
// inside one transaction that is always rolled back, so the database is left as it was.
//
//	go test -tags dbtest -run TestWiringKeepsDecimalQty -v ./services/sales_services/
func TestWiringKeepsDecimalQty(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	tx := initializers.DB.Begin()
	defer tx.Rollback()

	for _, sql := range []string{
		"ALTER TABLE dbo.tbl_trans_sales_project_wiring ALTER COLUMN qty decimal(18,4) NULL",
		"ALTER TABLE dbo.tbl_trans_sales_project_wiring ALTER COLUMN total_qty decimal(18,4) NULL",
		"ALTER TABLE dbo.z_tbl_trans_sales_project_wiring_at ALTER COLUMN qty decimal(18,4) NULL",
		"ALTER TABLE dbo.z_tbl_trans_sales_project_wiring_at ALTER COLUMN total_qty decimal(18,4) NULL",
	} {
		if err := tx.Exec(sql).Error; err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	// A new row, as an inserted wiring line is written.
	row := models.SalesProjectWiring{SalesProjectWiringContent: models.SalesProjectWiringContent{
		BasedId: 999999, Materials: "ECB To Controller", NumOfWiresSet: "3", NumOfQtySet: "1",
		DistanceTravelledSet: "10", AllowanceWireSet: "0.5", Qty: 31.5, NumOfSets: "1", TotalQty: 31.5,
		Cost: 100, TotalCost: 3150,
	}}
	if err := services.DbInsert(tx, &row); err != nil {
		t.Fatal(err)
	}

	var back models.SalesProjectWiring
	if err := tx.First(&back, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if back.Qty != 31.5 || back.TotalQty != 31.5 {
		t.Errorf("inserted: qty %v, total_qty %v, want 31.5 / 31.5", back.Qty, back.TotalQty)
	}

	// An update of that row, as a later edit is written: only the changed fields.
	sent := back
	sent.DistanceTravelledSet = "12.25"
	sent.Qty = 38.25
	sent.TotalQty = 76.5
	sent.NumOfSets = "2"
	changes := map[string]FieldChange[any]{
		"distance_travelled_set": {OldValue: "10", NewValue: "12.25"},
		"qty":                    {OldValue: 31.5, NewValue: 38.25},
		"total_qty":              {OldValue: 31.5, NewValue: 76.5},
		"num_of_sets":            {OldValue: "1", NewValue: "2"},
	}
	if err := writeChangedFields(tx, &sent, "id", back.ID, changes); err != nil {
		t.Fatal(err)
	}

	var updated models.SalesProjectWiring
	if err := tx.First(&updated, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Qty != 38.25 || updated.TotalQty != 76.5 {
		t.Errorf("updated: qty %v, total_qty %v, want 38.25 / 76.5", updated.Qty, updated.TotalQty)
	}
	t.Logf("insert kept %v / %v, update kept %v / %v (rolled back)", back.Qty, back.TotalQty, updated.Qty, updated.TotalQty)
}
