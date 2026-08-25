package models

// DebitMemoContent — header fields per spec §5.19/§12.6.
//
// A/P only - no customer counterpart, deliberately (§14.56, §14.100): "a
// debit against a customer - charging them more - is a new Sales Invoice,
// not a memo." Commits entirely on SAVE - no draft, no approval workflow
// (§12.6.3, §14.57).
//
// Field name is SupplierId/SupplierCode/SupplierName, not "customer" -
// the spec explicitly calls out a design-mockup error where the field was
// labelled CUSTOMER but populated with an S# ("the label is the error, not
// the data"). Do not repeat that mistake here.
type DebitMemoContent struct {
	SupplierId   uint   `gorm:"not null" json:"supplier_id"`
	SupplierCode string `gorm:"size:50" json:"supplier_code,omitempty"`
	SupplierName string `gorm:"size:255" json:"supplier_name,omitempty"`

	TransAmount float64 `json:"trans_amount"`

	// Required (§14.58). Fixed 5-value list per §5.19 (pur return, adj
	// twas, cancel chq, pur disc, exp cancel) - not present in §17 despite
	// §17 being described as authoritative; treated as fixed/inline per
	// the spec text until that gap is resolved, not promoted to an
	// editable Setup list on our own authority.
	ReasonCode string `gorm:"size:50;not null" json:"reason_code"`

	Currency      string `gorm:"size:10" json:"currency,omitempty"`
	LocationGroup string `gorm:"size:100" json:"location_group,omitempty"`
	DocDate       string `json:"doc_date,omitempty"`
	SalesPeriod   string `gorm:"size:20" json:"sales_period,omitempty"`
	RefDocNo      string `gorm:"size:50" json:"ref_doc_no,omitempty"`
	RefPoNo       string `gorm:"size:50" json:"ref_po_no,omitempty"`

	// Cached for listing/reporting - TransAmount minus the sum of every
	// ticked line's AmountApplied. Recompute from DebitMemoDetails on every
	// read that matters; §14.43 requires this to be 0 before the memo may
	// be finalized, enforced in the service layer at save time.
	UnappliedAmount float64 `json:"unapplied_amount"`

	DebitMemoDetails []DebitMemoDetails `gorm:"foreignKey:DebitMemoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"debit_memo_details,omitempty"`
}

type DebitMemo struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"not null;uniqueIndex:idx_tbl_trans_debit_memo_doc_no" json:"doc_no"`
	DebitMemoContent
}

func (DebitMemo) TableName() string {
	return "tbl_trans_debit_memo"
}

type DebitMemoAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DebitMemoContent
	At
}

func (DebitMemoAt) TableName() string {
	return "z_tbl_trans_debit_memo_at"
}

// DebitMemoDetailsContent — the apply table (§5.19 lines 1244-1257).
// Nothing applies until Apply is ticked (a required checkbox, not
// decorative); AmountApplied is only meant to be editable once ticked -
// enforced client-side, the service only checks the resulting sum.
//
// TargetDocType is one of "Invoice Receipt" / "Bulk Invoice Receipt" /
// "Credit Memo" - deliberately NOT "Miscellaneous Receiving", which is
// inherited from the previous system and explicitly out of scope (§15,
// §5.19): "Lightspeed has no such document and does not need one."
type DebitMemoDetailsContent struct {
	DebitMemoID uint `gorm:"not null;index" json:"debit_memo_id"`

	Apply         bool    `json:"apply"`
	TargetDocType string  `gorm:"size:30;not null" json:"target_doc_type"`
	TargetDocId   uint    `gorm:"not null" json:"target_doc_id"`
	TargetDocNo   string  `gorm:"size:50" json:"target_doc_no,omitempty"`
	DueDate       string  `json:"due_date,omitempty"`
	Total         float64 `json:"total"`
	OpenAmount    float64 `json:"open_amount"`
	AmountApplied float64 `json:"amount_applied"`

	// Computed (OpenAmount - AmountApplied), cached for display.
	Balance float64 `json:"balance"`
}

type DebitMemoDetails struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DebitMemoDetailsContent
}

func (DebitMemoDetails) TableName() string {
	return "tbl_trans_debit_memo_details"
}

type DebitMemoDetailsAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	DebitMemoDetailsContent
	At
}

func (DebitMemoDetailsAt) TableName() string {
	return "z_tbl_trans_debit_memo_details_at"
}

type DebitMemoBody struct {
	DebitMemo        DebitMemo          `json:"debit_memo"`
	DebitMemoDetails []DebitMemoDetails `json:"debit_memo_details"`
}

type DebitMemoGet struct {
	DebitMemo        []DebitMemo        `json:"debit_memo"`
	DebitMemoDetails []DebitMemoDetails `json:"debit_memo_details"`
}
