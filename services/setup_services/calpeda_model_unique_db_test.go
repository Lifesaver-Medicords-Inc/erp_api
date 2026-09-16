//go:build dbtest

package setup_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// Spec 4.2.2: Calpeda model names MUST be globally unique, other brands MAY repeat a model
// when the specs differ. Nothing enforced this before 2026-09-16 - the client only checked the
// catalogue year and the API only checked item_code - so this covers the new guard.
//
// Read-only: checkCalpedaModelUnique only counts, so the test asserts against data already in
// the database and writes nothing.
//
//	go test -tags dbtest -run TestCalpedaModel -v ./services/setup_services/
func TestCalpedaModelUniquenessIsEnforced(t *testing.T) {
	connectForTest(t)

	var calpeda models.Brand
	if err := initializers.DB.Where("code = ?", "CLP").Take(&calpeda).Error; err != nil {
		t.Skipf("no Calpeda brand in this database: %v", err)
	}

	var item models.Item
	if err := initializers.DB.Where("item_brand_id = ? AND ISNULL(item_model, '') <> ''", calpeda.ID).
		Order("id ASC").Take(&item).Error; err != nil {
		t.Skipf("no Calpeda item with a model: %v", err)
	}

	// A NEW item claiming an existing Calpeda model must be rejected.
	if err := checkCalpedaModelUnique(initializers.DB, calpeda.ID, item.ItemModel, 0); err == nil {
		t.Errorf("creating a second Calpeda item with model %q was allowed", item.ItemModel)
	}

	// The SAME item re-saving its own model must pass, or editing anything else on a Calpeda
	// item would be impossible.
	if err := checkCalpedaModelUnique(initializers.DB, calpeda.ID, item.ItemModel, item.ID); err != nil {
		t.Errorf("item %d could not re-save its own model %q: %v", item.ID, item.ItemModel, err)
	}

	// A model nobody holds must pass.
	if err := checkCalpedaModelUnique(initializers.DB, calpeda.ID, "ZZ-NOT-A-REAL-MODEL-2026", 0); err != nil {
		t.Errorf("an unused Calpeda model was rejected: %v", err)
	}

	// Blank model and blank brand are not this guard's business.
	if err := checkCalpedaModelUnique(initializers.DB, calpeda.ID, "   ", 0); err != nil {
		t.Errorf("a blank model should be left to the required-field checks: %v", err)
	}
	if err := checkCalpedaModelUnique(initializers.DB, 0, item.ItemModel, 0); err != nil {
		t.Errorf("no brand should skip the guard entirely: %v", err)
	}

	// Another brand reusing that same model must be allowed - 4.2.2 applies to Calpeda alone.
	var other models.Brand
	if err := initializers.DB.Where("code <> ?", "CLP").Order("id ASC").Take(&other).Error; err == nil {
		if err := checkCalpedaModelUnique(initializers.DB, other.ID, item.ItemModel, 0); err != nil {
			t.Errorf("brand %s was held to the Calpeda rule: %v", other.Code, err)
		}
	}

	var calpedaItems int64
	initializers.DB.Model(&models.Item{}).Where("item_brand_id = ?", calpeda.ID).Count(&calpedaItems)
	t.Logf("%s: guard checked against %d Calpeda items, sample model %q",
		os.Getenv("DB_NAME"), calpedaItems, item.ItemModel)
}
