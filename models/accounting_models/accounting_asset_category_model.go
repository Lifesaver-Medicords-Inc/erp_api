package accounting_models

import "github.com/pierceperado/smpc/models"

// AssetCategoryContent — the category level of the PP&E register (e.g.
// LAND, BUILDING, MACHINERY, OFFICE EQUIPMENT, VEHICLE), matching the
// asset-category breakdown the real SEC-filed financials show in their
// PP&E note. Not in SMPC_ERP_SPEC_*.md at all - this whole register is net
// new scope, found missing while building the Balance Sheet/Income
// Statement against that filing, not a spec-vs-code discrepancy.
type AssetCategoryContent struct {
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
}

type AssetCategory struct {
	ID uint `gorm:"primarykey" json:"id"`
	AssetCategoryContent
}

func (AssetCategory) TableName() string {
	return "tbl_setup_asset_category"
}

type AssetCategoryAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	AssetCategoryContent
	models.At
}

func (AssetCategoryAt) TableName() string {
	return "z_tbl_setup_asset_category_at"
}
