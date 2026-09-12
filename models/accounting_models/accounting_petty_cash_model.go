package accounting_models

import "github.com/pierceperado/smpc/models"

// Petty Cash Replenishment (spec 5.26). An imprest fund: fixed size, topped back
// up each cycle by exactly what was spent. One cycle is one request, and the next
// opens as a copy of the last with its rows cleared and the balance carried
// forward.
//
// Kept by the custodian (the A/P accountant) and approved by the COO.
//
// ACCOUNTABILITY is the computed result of a COUNTED cash on hand, never the
// other way round - the source workbook has that backwards and it is not
// reproduced here. The request cannot be submitted until
// CASH ON HAND + FOR REIMBURSEMENT equals the accountable fund (5.26, 2.9).
//
// No journal entry is posted: 12.7.4 records two unresolved defects in the source
// posting map - PETTY CASH PENDING is mapped to an account that does not exist,
// and the replenishment row debits a loss account instead of the fund. Under 12.8
// an unmapped logical name blocks the posting, so posting waits on management.
type PettyCashContent struct {
	DocNo      int    `json:"doc_no"`
	CutOffDate string `json:"cut_off_date"`

	// Carried forward from the previous cycle - explicitly, never hard-coded.
	OpeningFund float64 `json:"opening_fund"`
	// The fund the custodian answers for; what the tally check must land on.
	AccountableFund float64 `json:"accountable_fund"`
	// From the physical count. Not derived.
	CashOnHand float64 `json:"cash_on_hand"`
	// Non-replenishable deductions, entered as negatives.
	Deductions float64 `json:"deductions"`

	TotalExpense float64 `json:"total_expense"`
	// A single negative offset for purchases funded by encashing a cheque, so the
	// expense is documented but the cash is not reimbursed twice.
	Encashment       float64 `json:"encashment"`
	ForReimbursement float64 `json:"for_reimbursement"`
	Accountability   float64 `json:"accountability"`

	Status       string `json:"status"` // DRAFT | SUBMITTED | APPROVED
	PreparedBy   string `json:"prepared_by"`
	ApprovedBy   string `json:"approved_by"`
	ApprovedDate string `json:"approved_date"`
	Remarks      string `json:"remarks"`
}

type PettyCash struct {
	ID      uint               `gorm:"primarykey" json:"id"`
	Details []PettyCashDetails `gorm:"foreignKey:PettyCashId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"details"`
	PettyCashContent
}

func (PettyCash) TableName() string {
	return "tbl_accounting_petty_cash"
}

type PettyCashAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PettyCashContent
	models.At
}

func (PettyCashAt) TableName() string {
	return "z_tbl_accounting_petty_cash_at"
}

// One disbursement. PARTICULAR prints as "<requester>- <description>
// <references>"; the parts are stored separately so the sheet can be read and
// searched.
//
// A row with no document reference MUST be accepted (5.26): meals, unrequisitioned
// purchases and fares often have none, and a form that rejects them pushes real
// spending out of the record. The UI prompts; it never blocks.
type PettyCashDetailsContent struct {
	PettyCashId uint    `json:"petty_cash_id"`
	Date        string  `json:"date"`
	Requester   string  `json:"requester"`
	Description string  `json:"description"`
	Reference   string  `json:"reference"`
	Amount      float64 `json:"amount"`
	// True on the single offsetting negative line that cites the cheque and the
	// POs it covered.
	IsEncashment bool `json:"is_encashment"`
}

type PettyCashDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	PettyCashDetailsContent
}

func (PettyCashDetails) TableName() string {
	return "tbl_accounting_petty_cash_details"
}

type PettyCashDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PettyCashDetailsContent
	models.At
}

func (PettyCashDetailsAt) TableName() string {
	return "z_tbl_accounting_petty_cash_details_at"
}
