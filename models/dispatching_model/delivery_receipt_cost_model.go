package dispatching_models

import "github.com/pierceperado/smpc/models"

// User decision, 2026-09-03: this is now the ONE delivery-cost table. The logistics
// route's cost table (tbl_dispatching_logistics_route_cost) has been removed and its
// rows live here, so a cost is entered once and both the route form and the Delivery
// Receipt's DELIVERY COST block read the same row - §13.3, "identical to the DR's and
// the SO's, entered once".
//
// Two owners, both optional:
//   - RouteId            set when the row was entered on a logistics route
//   - DeliveryReceiptID  set when that route names a Delivery Receipt, or when the row
//                        was entered on the DR directly
//
// A route with no Delivery Receipt (an office-to-office trip, or a pick-up against a
// PO - §13.3) still records its costs; those rows simply carry RouteId alone. Zero
// means "no owner on that side": the existing delivery_receipt_id column is NOT NULL,
// so 0 is used as the sentinel rather than altering a live column to nullable, which
// GORM's additive AutoMigrate would not apply anyway.
type DeliveryReceiptCostContent struct {
	DeliveryReceiptID uint `gorm:"not null;default:0;index" json:"delivery_receipt_id"`
	RouteId           uint `gorm:"default:0;index" json:"route_id"`

	// COST TYPE is carried twice on purpose. The Delivery Receipt binds a combo to
	// cost_type_id (tbl_setup_calendar_cost_type); the logistics route form posts and
	// displays the plain name from its own hard-coded list. Storing both keeps either
	// screen working without a lookup on read, and the API keeps them consistent on
	// write (see NormalizeCostType).
	//
	// §17.5 says Cost Type is hard-coded and "not configurable in Setup", which the
	// id-based half contradicts. That discrepancy is reported separately and is not
	// settled here - holding both columns means whichever way it is decided, the
	// column that survives is already populated.
	CostTypeID uint   `gorm:"default:0" json:"cost_type_id"`
	CostType   string `gorm:"size:100" json:"cost_type,omitempty"`

	Description string  `gorm:"size:255" json:"description,omitempty"`
	Amount      float64 `gorm:"default:0" json:"amount"`
	Multiplier  float64 `gorm:"default:0" json:"multiplier"`
	TotalCost   float64 `gorm:"default:0" json:"total_cost"`

	// The route form uploads a single receipt and stores its path; the Delivery
	// Receipt attaches many files per cost line. Both are kept - ReceiptPath is what
	// the route grid round-trips, ReceiptFiles is what the DR's UPLOAD column manages.
	ReceiptPath  string        `gorm:"size:500" json:"receipt_path,omitempty"`
	ReceiptFiles []ReceiptFile `gorm:"foreignKey:DeliveryReceiptCostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"receipt_files,omitempty"`
}

type DeliveryReceiptCosts struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeliveryReceiptCostContent
}

func (DeliveryReceiptCosts) TableName() string {
	return "tbl_dispatching_delivery_receipt_costs"
}

type DeliveryReceiptCostsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DeliveryReceiptCostContent
	models.At
}

func (DeliveryReceiptCostsAt) TableName() string {
	return "z_tbl_dispatching_delivery_receipt_costs"
}
