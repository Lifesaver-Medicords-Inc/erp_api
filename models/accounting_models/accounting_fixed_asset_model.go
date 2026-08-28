package accounting_models

import "github.com/pierceperado/smpc/models"

// FixedAssetContent — one PP&E item. Depreciation is computed on the fly
// (straight-line, see fixed_asset_services) rather than stored as a running
// balance, the same "compute from source data at read time" approach this
// codebase already uses for Cost of Sales (FIFO lot consumption) rather
// than a periodic-inventory formula - no depreciation run/close step
// exists or is needed for this minimal register.
//
// Dates are free-text strings in "01/02/2006" (MM/dd/yyyy) form, matching
// every other date field in this codebase (tbl_trans_sales_order.doc_date
// and friends) - not a real DATE column.
type FixedAssetContent struct {
	// User-typed asset tag, not an auto-numbered document - PP&E entries are
	// low-volume, hand-entered master data, not a transactional document
	// series (unlike SO#/PO#/etc.).
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`

	// Denormalized at write time, same convention as ChartOfAccountContent's
	// AccountClass/ClassId pair - not a live join.
	CategoryId   uint   `gorm:"not null" json:"category_id"`
	CategoryName string `json:"category_name,omitempty"`

	Cost         float64 `json:"cost"`
	AcquiredDate string  `json:"acquired_date"`

	// Straight-line inputs. UsefulLifeYears == 0 marks a non-depreciable
	// asset (e.g. Land) - GetNetBookValue/GetAccumulatedDepreciation both
	// treat that as "never depreciates", not a divide-by-zero.
	UsefulLifeYears float64 `json:"useful_life_years"`
	SalvageValue    float64 `json:"salvage_value"`

	// "ACTIVE" | "DISPOSED". A disposed asset drops out of the Balance
	// Sheet/Income Statement rollups as of DisposedDate - it's excluded
	// entirely rather than depreciation-frozen, since a disposal removes
	// the asset from the books, it doesn't stop wearing it out on paper.
	Status       string `gorm:"size:20;not null;default:ACTIVE" json:"status"`
	DisposedDate string `json:"disposed_date,omitempty"`
}

type FixedAsset struct {
	ID uint `gorm:"primarykey" json:"id"`
	FixedAssetContent
}

func (FixedAsset) TableName() string {
	return "tbl_fixed_asset"
}

type FixedAssetAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	FixedAssetContent
	models.At
}

func (FixedAssetAt) TableName() string {
	return "z_tbl_fixed_asset_at"
}
