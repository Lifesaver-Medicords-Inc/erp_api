package models

// SalesReturnDetailsContent — line fields per spec §5.13.
//
// Every received unit gets exactly one destination: QtyForReplacement +
// QtyToStock + QtyForPurchaseReturn MUST equal QtyReceived (§5.13, §14 test
// #65). That sum check is enforced in the service layer at save time, not
// here - GORM has no cross-column CHECK constraint support worth relying on
// across the DB engines this could run against.
//
// UnitPrice and TotalCost come from the reference document and are
// read-only here by convention (never edited from this form) - §5.13, §14
// test #95. There are deliberately no discount fields: the line price
// already carries whatever discount/mark-up it was sold at (§8.3), and the
// header discounts on the original sale apply once against that sale's
// net_sales, not per return line (§12.6.2, §14 test #94).
type SalesReturnDetailsContent struct {
	SalesReturnID uint `gorm:"not null;index" json:"sales_return_id"`

	ItemID        uint   `gorm:"not null" json:"item_id"`
	ItemCode      string `gorm:"size:50" json:"item_code,omitempty"`
	Description   string `gorm:"size:255" json:"description,omitempty"`
	UnitOfMeasure string `gorm:"size:50" json:"unit_of_measure,omitempty"`

	// What the customer declared vs. what actually arrived. QtyDiscrepancy
	// is computed (QtyReturned - QtyReceived), never typed - stored here as
	// a cache only, recompute at the point of use (§5.13).
	QtyReturned    int `gorm:"not null" json:"qty_returned"`
	QtyReceived    int `gorm:"not null;default:0" json:"qty_received"`
	QtyDiscrepancy int `gorm:"not null;default:0" json:"qty_discrepancy"`

	// The three-way destination split. Mutually exclusive by unit, MUST sum
	// to QtyReceived (§5.13, §14 test #65).
	QtyForReplacement    int `gorm:"not null;default:0" json:"qty_for_replacement"`
	QtyToStock           int `gorm:"not null;default:0" json:"qty_to_stock"`
	QtyForPurchaseReturn int `gorm:"not null;default:0" json:"qty_for_purchase_return"`

	// Read-only, sourced from the reference document line (§5.13, §14 test
	// #95) - the discounted/marked-up price the unit was actually sold at.
	UnitPrice float64 `json:"unit_price"`
	TotalCost float64 `json:"total_cost"`

	ReasonForReturn string `gorm:"type:text" json:"reason_for_return,omitempty"`
}

type SalesReturnDetails struct {
	ID uint `gorm:"primaryKey" json:"id"`
	SalesReturnDetailsContent
}

func (SalesReturnDetails) TableName() string {
	return "tbl_trans_sales_return_details"
}

type SalesReturnDetailsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesReturnDetailsContent
	At
}

func (SalesReturnDetailsAt) TableName() string {
	return "z_tbl_trans_sales_return_details_at"
}
