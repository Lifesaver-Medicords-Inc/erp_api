package initializers

import (
	"fmt"

	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
)

// SeedCalendarCostTypes ensures the fixed Delivery Cost "COST TYPE" list
// (LABOR/VEHICLE/FUEL/TOLL GATE/INSURANCE/PENALTY/OTHERS, per the Delivery
// Cost grid mock-up) exists in tbl_setup_calendar_cost_type. Insert-if-
// missing, keyed on Code - same idempotent, restart-safe convention as
// SeedAccessModules. Never edits or deletes an existing row, so a rename
// or addition made later from the Cost Type setup screen itself is never
// clobbered on the next restart.
func SeedCalendarCostTypes() {
	defaults := []struct{ Code, Name string }{
		{"LABOR", "Labor"},
		{"VEHICLE", "Vehicle"},
		{"FUEL", "Fuel"},
		{"TOLL_GATE", "Toll Gate"},
		{"INSURANCE", "Insurance"},
		{"PENALTY", "Penalty"},
		{"OTHERS", "Others"},
	}

	inserted := 0
	for _, d := range defaults {
		var existing dispatching_models.CalendarCostTypeModel
		if DB.Where("code = ?", d.Code).First(&existing).Error == nil {
			continue // already seeded
		}

		row := dispatching_models.CalendarCostTypeModel{
			CalendarCostTypeContent: dispatching_models.CalendarCostTypeContent{
				Code: d.Code,
				Name: d.Name,
			},
		}
		if err := DB.Create(&row).Error; err != nil {
			fmt.Println("❌ SeedCalendarCostTypes: failed inserting", d.Code, "|", err)
			continue
		}
		inserted++
	}

	if inserted > 0 {
		fmt.Printf("✅ SeedCalendarCostTypes: inserted %d new cost type row(s)\n", inserted)
	}
}

// SeedDefaultProjectQuotationTemplate is a safety net for a brand-new
// environment ONLY - it does nothing unless tbl_trans_sales_project_template
// is completely empty (confirmed with the user: this must never touch a DB
// that already has real templates, which is the normal case once anyone has
// used Project Quotation at all).
//
// When it does fire, it prefers reusing whatever master data already
// exists: an Item Entry needs an Item Name/Class/Brand/Unit of Measure to
// point at, and on a genuinely fresh DB none of those exist yet either -
// that dependency chain is the "tricky" part. Each level is seeded only if
// its own table is empty; if even one real Item Entry already exists, that
// one is reused instead of fabricating another, so this never pollutes a
// real catalog with a placeholder. Everything this function creates is
// named/coded "DEFAULT"/"PLACEHOLDER" so it can never be mistaken for real
// data and is easy to find and delete once real templates exist.
func SeedDefaultProjectQuotationTemplate() {
	var templateCount int64
	if err := DB.Model(&models.SalesProjectTemplate{}).Count(&templateCount).Error; err != nil {
		fmt.Println("❌ SeedDefaultProjectQuotationTemplate: failed checking existing templates:", err)
		return
	}
	if templateCount > 0 {
		return // real templates already exist - never touch this DB
	}

	itemId, err := ensureDefaultItemForSeeding()
	if err != nil {
		fmt.Println("❌ SeedDefaultProjectQuotationTemplate: failed resolving a default item:", err)
		return
	}

	template := models.SalesProjectTemplate{
		SalesProjectTemplateContent: models.SalesProjectTemplateContent{
			TemplateName: "DEFAULT STARTER TEMPLATE (PLACEHOLDER)",
		},
	}
	if err := DB.Create(&template).Error; err != nil {
		fmt.Println("❌ SeedDefaultProjectQuotationTemplate: failed creating default template:", err)
		return
	}

	child := models.SalesProjectTemplateChild{
		ParentID: template.TemplateID,
		SalesProjectTemplateChildContent: models.SalesProjectTemplateChildContent{
			ParentId: template.TemplateID,
			ItemID:   itemId,
			Level:    1,
		},
	}
	if err := DB.Create(&child).Error; err != nil {
		fmt.Println("❌ SeedDefaultProjectQuotationTemplate: failed creating default template child:", err)
		return
	}

	fmt.Println("✅ SeedDefaultProjectQuotationTemplate: no templates existed - created one placeholder template + item so Project Quotation's template picker isn't blank on a fresh DB")
}

// ensureDefaultItemForSeeding returns an existing Item Entry's id if any
// exist, otherwise fabricates the minimal chain (Name/Class/Brand/Unit of
// Measure -> Item) needed to create exactly one, all clearly tagged as a
// placeholder. Only called from SeedDefaultProjectQuotationTemplate, whose
// own empty-templates guard already limits this to a genuinely fresh DB.
func ensureDefaultItemForSeeding() (uint, error) {
	var existingItem models.Item
	if err := DB.First(&existingItem).Error; err == nil {
		return existingItem.ID, nil // reuse real catalog data - never fabricate a second item
	}

	nameId, err := ensureDefaultLookupRow(&models.Name{}, "tbl_setup_item_name", "DEFAULT", "Default Item Name (Placeholder)")
	if err != nil {
		return 0, err
	}
	classId, err := ensureDefaultLookupRow(&models.Class{}, "tbl_setup_item_class", "DEFAULT", "Default (Placeholder)")
	if err != nil {
		return 0, err
	}
	brandId, err := ensureDefaultLookupRow(&models.Brand{}, "tbl_setup_item_brand", "DEFAULT", "Default (Placeholder)")
	if err != nil {
		return 0, err
	}
	uomId, err := ensureDefaultLookupRow(&models.UnitMeasurement{}, "tbl_setup_item_unit_measurement", "UNIT", "Unit (Placeholder)")
	if err != nil {
		return 0, err
	}

	stopSelling := false
	item := models.Item{
		ItemContent: models.ItemContent{
			ItemNameId:          nameId,
			ItemClassId:         classId,
			ItemBrandId:         brandId,
			UnitOfMeasureId:     uomId,
			ItemModel:           "DEFAULT-PLACEHOLDER",
			ItemCode:            "DEFAULT-PLACEHOLDER",
			ItemTangibilityType: "TANGIBLE",
			IsStopSelling:       &stopSelling,
			Price:               0,
		},
	}
	if err := DB.Create(&item).Error; err != nil {
		return 0, err
	}
	return item.ID, nil
}

// ensureDefaultLookupRow returns the first existing row's id for whichever
// of Name/Class/Brand/UnitOfMeasure model is passed in, or creates exactly
// one placeholder row (Code/Name as given) if that table is completely
// empty. All four share the same {ID, Code, Name} shape, but are distinct
// Go types with no common interface for Code/Name, so the model is only
// used to pick the right table via reflection-free type switch.
func ensureDefaultLookupRow(model interface{}, tableName string, code string, name string) (uint, error) {
	type lookupRow struct {
		ID   uint
		Code string
		Name string
	}
	var existing lookupRow
	if err := DB.Table(tableName).Order("id").Limit(1).Scan(&existing).Error; err == nil && existing.ID != 0 {
		return existing.ID, nil
	}

	switch model.(type) {
	case *models.Name:
		row := models.Name{Code: code, NameContent: models.NameContent{Name: name}}
		if err := DB.Create(&row).Error; err != nil {
			return 0, err
		}
		return row.ID, nil
	case *models.Class:
		row := models.Class{Code: code, ClassContent: models.ClassContent{Name: name}}
		if err := DB.Create(&row).Error; err != nil {
			return 0, err
		}
		return row.ID, nil
	case *models.Brand:
		row := models.Brand{Code: code, BrandContent: models.BrandContent{Name: name}}
		if err := DB.Create(&row).Error; err != nil {
			return 0, err
		}
		return row.ID, nil
	case *models.UnitMeasurement:
		row := models.UnitMeasurement{Code: code, UnitMeasurementContent: models.UnitMeasurementContent{Name: name}}
		if err := DB.Create(&row).Error; err != nil {
			return 0, err
		}
		return row.ID, nil
	default:
		return 0, fmt.Errorf("ensureDefaultLookupRow: unsupported model type for table %s", tableName)
	}
}
