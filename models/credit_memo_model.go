package models

// CreditMemoContent — header fields per spec §5.18, §12.6.
//
// One document serves both A/P and A/R. Direction (payable up vs
// receivable down) is derived from PartnerType, which is NEVER a value the
// client is free to pick arbitrarily - see CreateCreditMemo's own comment
// for how it's actually constrained (the creating module fixes it, and the
// service validates the partner genuinely holds that BPI entity type,
// since one partner can legitimately be both a customer and a supplier at
// once - confirmed via tbl_bpi_entity, which allows both rows for the same
// bpi_general_id).
type CreditMemoContent struct {
	PartnerId   uint   `gorm:"not null" json:"partner_id"`
	PartnerCode string `gorm:"size:50" json:"partner_code,omitempty"`
	PartnerName string `gorm:"size:255" json:"partner_name,omitempty"`

	// "Supplier" | "Customer" - sets the direction. Supplier => payable up.
	// Customer => receivable down (§14.98).
	PartnerType string `gorm:"size:20;not null" json:"partner_type"`

	TransAmount float64 `json:"trans_amount"`

	// Required on both sides (§14.58).
	ReasonCode string `gorm:"size:50;not null" json:"reason_code"`

	Currency      string `gorm:"size:10" json:"currency,omitempty"`
	LocationGroup string `gorm:"size:100" json:"location_group,omitempty"`
	DocDate       string `json:"doc_date,omitempty"`
	SalesPeriod   string `gorm:"size:20" json:"sales_period,omitempty"`

	// Customer side only - the originating Sales Return and the SI it
	// credits (inherited from the SRT's own reference doc). Both blank for
	// a credit raised with no return at all (§12.6.4 - overbilling, price
	// correction, goodwill; "no one should ever have to record a fictitious
	// goods return in order to issue a credit").
	RefSrtId uint   `json:"ref_srt_id,omitempty"`
	RefSrtNo string `gorm:"size:50" json:"ref_srt_no,omitempty"`
	RefSiId  uint   `json:"ref_si_id,omitempty"`
	RefSiNo  string `gorm:"size:50" json:"ref_si_no,omitempty"`

	// Supplier side only - ticking DmRefund + RefDmNo refunds that DM's
	// unapplied debit. §14.100: a customer CM MUST NOT offer these.
	DmRefund *bool  `json:"dm_refund,omitempty"`
	RefDmId  uint   `json:"ref_dm_id,omitempty"`
	RefDmNo  string `gorm:"size:50" json:"ref_dm_no,omitempty"`

	// Approval gate - customer side ONLY (§5.18, §3.3). A supplier CM
	// commits on SAVE and never touches these; IsApproved stays false and
	// unused for that row rather than being repurposed to mean "posted".
	IsApproved     bool   `json:"is_approved"`
	ApprovedByID   uint   `json:"approved_by_id,omitempty"`
	ApprovedByName string `gorm:"size:255" json:"approved_by_name,omitempty"`
	ApprovalDate   string `json:"approval_date,omitempty"`

	// AppliedByDm: set true once a Debit Memo's apply line fully consumes
	// this (supplier) Credit Memo (§12.6.3: "a DM's save also updates every
	// account it was applied against"). Purely informational now, same as
	// InvoiceReceipt/BulkInvoiceReceipt's own ApVoucher flag - see
	// debit_memo_services.applyToTargetDocuments's doc comment. What
	// actually gates whether a CM still has room for a Debit Memo to apply
	// against it is OpenAmount below.
	AppliedByDm bool `json:"applied_by_dm"`

	// OpenAmount: TransAmount minus every Debit Memo detail already applied
	// against this CM (services.ComputeCreditMemoOpenAmount) - computed
	// live on read, never stored, same reasoning as the IR/Bulk IR side.
	// Populated by CreditMemoService.GetCreditMemo for a supplier CM only
	// (the only side a Debit Memo can target, §5.19); left at its zero
	// value for a customer CM, which has no such consumer.
	OpenAmount float64 `gorm:"-" json:"open_amount,omitempty"`
}

type CreditMemo struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"not null;uniqueIndex:idx_tbl_trans_credit_memo_doc_no" json:"doc_no"`
	CreditMemoContent
}

func (CreditMemo) TableName() string {
	return "tbl_trans_credit_memo"
}

type CreditMemoAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	CreditMemoContent
	At
}

func (CreditMemoAt) TableName() string {
	return "z_tbl_trans_credit_memo_at"
}

type CreditMemoBody struct {
	CreditMemo CreditMemo `json:"credit_memo"`
}

type CreditMemoGet struct {
	CreditMemo []CreditMemo `json:"credit_memo"`
}
