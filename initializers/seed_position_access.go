package initializers

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pierceperado/smpc/models"
)

// positionAccessSeedJSON maps each department in the access-level workbook's nine
// "FOR <DEPARTMENT>" menus to the page codes that back them, resolved against the catalogue
// SeedAccessModules loads (see seed_data/access_modules_seed.json). Pages only: the buttons
// on a granted page are worked out here, so adding a button to the catalogue does not mean
// revisiting this file.
//
//go:embed seed_data/position_access_seed.json
var positionAccessSeedJSON []byte

// What a button is attached to: buttons share their page's app, module and submodule.
type pageKey struct{ app, module, submodule string }

type positionAccessSeedEntry struct {
	Department string   `json:"department"`
	Positions  []string `json:"positions"`
	Pages      []string `json:"pages"`
	// Granted by code because the blanket rule below deliberately withholds them: these
	// are the destructive/approval buttons (spec 3.3 gives Approve and Cancel Order to the
	// Sales Manager and the CBDO, for instance).
	SensitiveButtons []string `json:"sensitive_buttons"`
}

// SeedPositionAccess gives a position its starting set of access codes.
//
// It only ever fills in a position that has NO grants at all. A position with even one row
// is one somebody has already set up through Admin's Access Control screen, and re-adding
// codes there would quietly undo a deliberate removal - access is granted by hand in this
// system (see CLAUDE.md), and a redeploy must not be a way to hand it back. Nothing is ever
// deleted here.
//
// A granted page carries the non-sensitive buttons on it, so a screen a position can open is
// a screen it can actually use. Sensitive buttons - Delete, Approve, Cancel Order, Finalize -
// are withheld and only granted where the seed names them.
func SeedPositionAccess() {
	var entries []positionAccessSeedEntry
	if err := json.Unmarshal(positionAccessSeedJSON, &entries); err != nil {
		fmt.Println("❌ SeedPositionAccess: failed parsing embedded seed JSON:", err)
		return
	}

	seedMissingPositions()

	var catalogue []models.AccessModuleModel
	if err := DB.Find(&catalogue).Error; err != nil {
		fmt.Println("❌ SeedPositionAccess: failed reading the access module catalogue:", err)
		return
	}

	// Buttons live under the page sharing their app, module and submodule.
	pageOf := map[string]pageKey{}
	buttonsOf := map[pageKey][]models.AccessModuleModel{}

	for _, row := range catalogue {
		key := pageKey{row.AppName, row.Module, row.Submodule}
		if row.Kind == "page" {
			pageOf[row.Code] = key
			continue
		}
		buttonsOf[key] = append(buttonsOf[key], row)
	}

	for _, entry := range entries {
		wanted := map[string]bool{}

		for _, code := range entry.Pages {
			key, known := pageOf[code]
			if !known {
				// A page named in the seed but missing from the catalogue: report it
				// rather than granting a code nothing will ever check.
				fmt.Println("⚠️  SeedPositionAccess:", entry.Department, "- no such page in the catalogue:", code)
				continue
			}

			wanted[code] = true
			for _, button := range buttonsOf[key] {
				if !button.IsSensitive {
					wanted[button.Code] = true
				}
			}
		}

		for _, code := range entry.SensitiveButtons {
			wanted[code] = true
		}

		for _, name := range entry.Positions {
			seedOnePosition(name, entry.Department, wanted, pageOf)
		}
	}
}

// Two of the workbook's nine departments have no position to grant to: the dispatcher and
// the AR/AP cashier. Created here so the seed below has somewhere to land - create-only, and
// only when nothing of that name exists, so a position renamed through Admin's Position Setup
// is never duplicated or overwritten.
func seedMissingPositions() {
	missing := []models.PositionModel{
		{Code: "DISPATCH", PositionContent: models.PositionContent{Name: "Dispatcher"}},
		{Code: "CASHIER", PositionContent: models.PositionContent{Name: "AR/AP Cashier"}},
	}

	for _, position := range missing {
		// Find, not First: a missing row is the normal case here, and First logs it as an
		// error on every startup.
		var existing []models.PositionModel
		if err := DB.Where("LTRIM(RTRIM(LOWER(name))) = ? OR code = ?",
			strings.ToLower(position.Name), position.Code).Limit(1).Find(&existing).Error; err != nil {
			fmt.Println("❌ SeedPositionAccess: failed looking up position", position.Name, "|", err)
			continue
		}
		if len(existing) > 0 {
			continue
		}

		if err := DB.Create(&position).Error; err != nil {
			fmt.Println("❌ SeedPositionAccess: failed creating position", position.Name, "|", err)
			continue
		}

		fmt.Println("✅ SeedPositionAccess: created position", position.Name)
	}
}

func seedOnePosition(name, department string, codes map[string]bool, pages map[string]pageKey) {
	// Position names differ slightly between databases ("Purchaser" on one, "Purschasing"
	// on another), which is why the seed lists every spelling it knows and matches without
	// regard to case or surrounding spaces. Find, not First, to keep a position this
	// database simply does not have out of the error log.
	var found []models.PositionModel
	if err := DB.Where("LTRIM(RTRIM(LOWER(name))) = ?", strings.ToLower(strings.TrimSpace(name))).
		Limit(1).Find(&found).Error; err != nil {
		fmt.Println("❌ SeedPositionAccess: failed looking up position", name, "|", err)
		return
	}
	if len(found) == 0 {
		return // not a position in this database - nothing to do
	}
	position := found[0]

	// Counted against the PAGES it can open, not its grants as a whole. A position holding
	// nothing but a couple of button codes - the Sales Manager and the CBDO each had just
	// Approve and Cancel Order - can open no screen at all, so its navigation would come up
	// empty; that is an un-set-up position, not a curated one. Any position that can already
	// open something is left exactly as it is.
	var granted []string
	if err := DB.Model(&models.PositionAccessModel{}).
		Where("position_id = ?", position.ID).Pluck("code", &granted).Error; err != nil {
		fmt.Println("❌ SeedPositionAccess: failed reading grants for", name, "|", err)
		return
	}

	for _, code := range granted {
		if _, isPage := pages[code]; isPage {
			return // already set up - see the note on SeedPositionAccess
		}
	}

	// Its existing button grants stay; only what it is missing is added.
	held := map[string]bool{}
	for _, code := range granted {
		held[code] = true
	}

	rows := make([]models.PositionAccessModel, 0, len(codes))
	for code := range codes {
		if held[code] {
			continue
		}

		rows = append(rows, models.PositionAccessModel{
			PositionAccessContent: models.PositionAccessContent{
				PositionId: position.ID,
				Code:       code,
			},
		})
	}

	if len(rows) == 0 {
		return
	}

	if err := DB.CreateInBatches(rows, 200).Error; err != nil {
		fmt.Println("❌ SeedPositionAccess: failed granting", name, "|", err)
		return
	}

	fmt.Printf("✅ SeedPositionAccess: %s (%s) granted %d codes\n", name, department, len(rows))
}
