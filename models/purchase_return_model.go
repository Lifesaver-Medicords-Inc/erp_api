package models

// PurchaseReturnContent — header fields per spec §5.8.
//
// Requires a supplier + a reference Invoice Receipt (never a PO) - the IR
// records what actually arrived and what will be paid, and a return can
// happen even when payment preceded delivery.
//
// Created two ways: (a) directly by a Purchaser regardless of payment
// state, or (b) auto-generated from a Sales Return line carrying a
// QtyForPurchaseReturn (§5.13, §12.6.1). SourceSalesReturnID/DetailsID are
// only set for path (b), so the resolution chain (SRT → SO → matching PO
// line → IR) stays traceable instead of re-derived every time it's read.
type PurchaseReturnContent struct {
	SupplierID   uint   `gorm:"not null" json:"supplier_id"`
	SupplierCode string `gorm:"size:50" json:"supplier_code,omitempty"`
	SupplierName string `gorm:"size:255" json:"supplier_name,omitempty"`
	TinNo        string `gorm:"size:50" json:"tin_no,omitempty"`
	Address      string `gorm:"size:255" json:"address,omitempty"`

	// The reference document is the Invoice Receipt, not the PO (§5.8).
	RefIRID uint   `gorm:"not null" json:"ref_ir_id"`
	RefIRNo string `gorm:"size:50" json:"ref_ir_no,omitempty"`

	// "Return with Debit Memo" | "Return without Debit Memo" (§5.8). With
	// = a receipt exists, so a Debit Memo is auto-created against the
	// supplier's account. Without = no receipt, no DM needed. The DM
	// itself is a separate model/service (item 1.8) - not built here.
	ReturnType string `gorm:"size:30;not null" json:"return_type"`

	// Populated only once the "with Debit Memo" type actually produces one.
	RefDmID uint   `json:"ref_dm_id,omitempty"`
	RefDmNo string `gorm:"size:50" json:"ref_dm_no,omitempty"`

	// Set only when this PRT was auto-generated from a Sales Return line's
	// QtyForPurchaseReturn (§12.6.1). Nil/zero on a purchaser-initiated
	// return. SourceSalesReturnDetailsID pins which specific SRT line
	// caused this row, since one SRT can spawn several PRT lines across
	// different suppliers.
	SourceSalesReturnID        uint `json:"source_sales_return_id,omitempty"`
	SourceSalesReturnDetailsID uint `json:"source_sales_return_details_id,omitempty"`

	Remarks string `gorm:"type:text" json:"remarks,omitempty"`

	PurchaseReturnDetails []PurchaseReturnDetails `gorm:"foreignKey:PurchaseReturnID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchase_return_details,omitempty"`
}

type PurchaseReturn struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"not null;uniqueIndex:idx_tbl_trans_purchase_return_doc_no" json:"doc_no"`
	PurchaseReturnContent
}

func (PurchaseReturn) TableName() string {
	return "tbl_trans_purchase_return"
}

type PurchaseReturnAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchaseReturnContent
	At
}

func (PurchaseReturnAt) TableName() string {
	return "z_tbl_trans_purchase_return_at"
}
