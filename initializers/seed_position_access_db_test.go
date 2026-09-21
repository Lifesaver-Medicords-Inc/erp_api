//go:build dbtest

package initializers

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/models"
)

// Runs SeedPositionAccess against the database DB_NAME points at and reports what each
// position ended up with. Writes: it grants positions that have no access at all, which is
// the whole point of the seeder - run it against a rehearsal database first.
//
//	go test -tags dbtest -run TestSeedPositionAccess -v ./initializers/
func TestSeedPositionAccess(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir(".."); err != nil {
			t.Fatal(err)
		}
	}
	LoadEnv()
	ConnectDb()

	var before int64
	if err := DB.Model(&models.PositionAccessModel{}).Count(&before).Error; err != nil {
		t.Fatal(err)
	}

	SeedPositionAccess()

	type row struct {
		Name    string
		Granted int64
	}
	var rows []row
	if err := DB.Raw(`SELECT p.name AS name, COUNT(a.id) AS granted
	                  FROM tbl_position p
	                  LEFT JOIN tbl_position_access a ON a.position_id = p.id
	                  GROUP BY p.name ORDER BY p.name`).Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}

	var after int64
	if err := DB.Model(&models.PositionAccessModel{}).Count(&after).Error; err != nil {
		t.Fatal(err)
	}

	t.Logf("%s: %d grants before, %d after", os.Getenv("DB_NAME"), before, after)
	for _, r := range rows {
		t.Logf("   %-40s %4d", r.Name, r.Granted)
		if r.Granted == 0 {
			t.Errorf("%s still has no access at all", r.Name)
		}
	}

	// Running it twice must not double anything up - it only fills positions that have
	// nothing, so the second pass has to be a no-op.
	SeedPositionAccess()

	var again int64
	if err := DB.Model(&models.PositionAccessModel{}).Count(&again).Error; err != nil {
		t.Fatal(err)
	}
	if again != after {
		t.Errorf("second run changed the grant count: %d -> %d", after, again)
	}
}
