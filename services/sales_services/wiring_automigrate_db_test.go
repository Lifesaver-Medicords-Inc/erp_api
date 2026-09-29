//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// Does the API's own start-up migration (AutoMigrate, as migrateSalesProject runs it) widen
// the wiring qty / total_qty columns from bigint to decimal(18,4) by itself, so no script is
// needed? And is a second start, with the columns already decimal, a no-op? Rolled back.
//
//	go test -tags dbtest -run TestWiringAutoMigrate -v ./services/sales_services/
func TestWiringAutoMigrate(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()

	tx := initializers.DB.Begin()
	defer tx.Rollback()

	types := func(label string) map[string]string {
		rows := []struct {
			TableName  string
			ColumnName string
			TypeNow    string
		}{}
		tx.Raw(`SELECT OBJECT_NAME(c.object_id) AS table_name, c.name AS column_name,
		               TYPE_NAME(c.user_type_id) + '(' + CAST(c.precision AS varchar(5)) + ',' + CAST(c.scale AS varchar(5)) + ')' AS type_now
		          FROM sys.columns c
		         WHERE c.object_id IN (OBJECT_ID('dbo.tbl_trans_sales_project_wiring'), OBJECT_ID('dbo.z_tbl_trans_sales_project_wiring_at'))
		           AND c.name IN ('qty', 'total_qty')`).Scan(&rows)
		out := map[string]string{}
		for _, r := range rows {
			out[r.TableName+"."+r.ColumnName] = r.TypeNow
			t.Logf("%-8s %s.%s = %s", label, r.TableName, r.ColumnName, r.TypeNow)
		}
		return out
	}

	types("before")

	if err := tx.AutoMigrate(&models.SalesProjectWiring{}, &models.SalesProjectWiringAt{}); err != nil {
		t.Fatalf("first start-up migration: %v", err)
	}
	after := types("after")
	for col, typ := range after {
		if typ != "decimal(18,4)" {
			t.Errorf("%s is %s after the start-up migration, want decimal(18,4)", col, typ)
		}
	}

	if err := tx.AutoMigrate(&models.SalesProjectWiring{}, &models.SalesProjectWiringAt{}); err != nil {
		t.Fatalf("second start-up migration (columns already decimal): %v", err)
	}
	types("again")
}
