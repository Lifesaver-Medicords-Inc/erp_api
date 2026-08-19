package initializers

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/pierceperado/smpc/models"
)

// accessModulesSeedJSON is the full catalog generated from
// SMPC_User_Access_Level_List.xlsx (168 screens / 797 total page+button entries across
// all 6 apps). See seed_data/access_modules_seed.json for the raw data and
// SMPC_User_Access_Level_List_PROPOSED_CODES.xlsx (delivered separately) for the
// human-readable version this was generated from.
//
//go:embed seed_data/access_modules_seed.json
var accessModulesSeedJSON []byte

type accessModuleSeedRow struct {
	AppName     string `json:"app_name"`
	Module      string `json:"module"`
	Submodule   string `json:"submodule"`
	Button      string `json:"button"`
	Code        string `json:"code"`
	Kind        string `json:"kind"`
	IsSensitive bool   `json:"is_sensitive"`
}

// SeedAccessModules loads the catalog once at startup. It's an upsert keyed on
// (app_name, code, button) - safe to run on every restart: rows that already exist are
// left untouched (so any Save from the Access Control screen that flips is_active isn't
// clobbered on the next deploy), rows missing from the DB but present in the seed file
// are inserted, and nothing is ever deleted here (removing a screen from the catalog is
// a manual/deliberate action, not an automatic side effect of a redeploy).
func SeedAccessModules() {
	var rows []accessModuleSeedRow
	if err := json.Unmarshal(accessModulesSeedJSON, &rows); err != nil {
		fmt.Println("❌ SeedAccessModules: failed parsing embedded seed JSON:", err)
		return
	}

	inserted := 0
	for _, r := range rows {
		var existing models.AccessModuleModel
		tx := DB.Where("app_name = ? AND code = ? AND button = ?", r.AppName, r.Code, r.Button).
			First(&existing)

		if tx.Error == nil {
			continue // already seeded
		}

		row := models.AccessModuleModel{
			AccessModuleContent: models.AccessModuleContent{
				AppName:     r.AppName,
				Module:      r.Module,
				Submodule:   r.Submodule,
				Button:      r.Button,
				Code:        r.Code,
				Kind:        r.Kind,
				IsSensitive: r.IsSensitive,
				IsActive:    true,
			},
		}

		if err := DB.Create(&row).Error; err != nil {
			fmt.Println("❌ SeedAccessModules: failed inserting", r.AppName, r.Code, r.Button, "|", err)
			continue
		}
		inserted++
	}

	if inserted > 0 {
		fmt.Printf("✅ SeedAccessModules: inserted %d new catalog row(s) (of %d total in seed file)\n", inserted, len(rows))
	}
}
