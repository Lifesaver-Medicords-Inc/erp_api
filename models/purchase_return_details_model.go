package models

// PurchaseReturnDetailsContent — line fields per spec §5.8/§12.6.1.
//
// RefIRDetailsID pins the return to a specific Invoice Receipt line, not
// just the IR header - one IR may span several POs, so the item list can
// too, and matching at the header level would let a return apply against
// the wrong PO/supplier. This is the exact mistake §5.8 step 3 warns
// about ("Match at line level, never at PO level... checking this is not
// optional, it is where the mistake happens") when this row is produced by
// the Sales-Return auto-generation path.
type PurchaseReturnDetailsContent struct {
	PurchaseReturnID uint `gorm:"not null;index" json:"purchase_return_id"`

	RefIRDetailsID uint `gorm:"not null" json:"ref_ir_details_id"`

	ItemID        uint   `gorm:"not null" json:"item_id"`
	ItemCode      string `gorm:"size:50" json:"item_code,omitempty"`
	Description   string `gorm:"size:255" json:"description,omitempty"`
	UnitOfMeasure string `gorm:"size:50" json:"unit_of_measure,omitempty"`

	Qty int `gorm:"not null" json:"qty"`

	// Auto-fills from the matched IR line; not user-entered.
	UnitCost float64 `json:"unit_cost"`

	Reason string `gorm:"type:text" json:"reason,omitempty"`
}

type PurchaseReturnDetails struct {
	ID uint `gorm:"primaryKey" json:"id"`
	PurchaseReturnDetailsContent
}

func (PurchaseReturnDetails) TableName() string {
	return "tbl_trans_purchase_return_details"
}

type PurchaseReturnDetailsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchaseReturnDetailsContent
	At
}

func (PurchaseReturnDetailsAt) TableName() string {
	return "z_tbl_trans_purchase_return_details_at"
}
