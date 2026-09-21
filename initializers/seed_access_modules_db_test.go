//go:build dbtest

package initializers

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/models"
)

// Runs SeedAccessModules against the database DB_NAME points at, so catalogue rows added to
// seed_data/access_modules_seed.json land without waiting for an API restart. Insert-only:
// it never rewrites or removes a row that is already there.
//
//	go test -tags dbtest -run TestSeedAccessModules -v ./initializers/
func TestSeedAccessModules(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir(".."); err != nil {
			t.Fatal(err)
		}
	}
	LoadEnv()
	ConnectDb()

	var before int64
	if err := DB.Model(&models.AccessModuleModel{}).Count(&before).Error; err != nil {
		t.Fatal(err)
	}

	SeedAccessModules()

	var after int64
	if err := DB.Model(&models.AccessModuleModel{}).Count(&after).Error; err != nil {
		t.Fatal(err)
	}

	t.Logf("%s: %d catalogue rows before, %d after", os.Getenv("DB_NAME"), before, after)

	// The two entries the inventory sidebar has always had without a code to grant.
	for _, code := range []string{"Purchasing.Purchase Return", "Inventory.Production Report"} {
		var found int64
		if err := DB.Model(&models.AccessModuleModel{}).
			Where("code = ? AND app_name = ?", code, "Inventory App").Count(&found).Error; err != nil {
			t.Fatal(err)
		}
		if found == 0 {
			t.Errorf("%s is missing from the catalogue after seeding", code)
		}
	}

	// A second pass must add nothing.
	SeedAccessModules()

	var again int64
	if err := DB.Model(&models.AccessModuleModel{}).Count(&again).Error; err != nil {
		t.Fatal(err)
	}
	if again != after {
		t.Errorf("second run changed the catalogue count: %d -> %d", after, again)
	}
}
