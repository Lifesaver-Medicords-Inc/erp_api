package models

// SalesReturnContent — header fields per spec §5.13. A Sales Return is a
// customer document created by A/R and requires Sales Manager approval
// before anything takes effect: no stock movement and no auto-generated
// Purchase Return happen until IsApproved is set (§5.13, §12.6.2).
type SalesReturnContent struct {
	CustomerID   uint   `gorm:"not null" json:"customer_id"`
	CustomerCode string `gorm:"size:50" json:"customer_code,omitempty"`
	CustomerName string `gorm:"size:255" json:"customer_name,omitempty"`
	TinNo        string `gorm:"size:50" json:"tin_no,omitempty"`
	Address      string `gorm:"size:255" json:"address,omitempty"`

	// RefDocType MUST be chosen before anything else - it decides which
	// document's lines may be pulled in (§5.13). One of "Sales Invoice" /
	// "Delivery Receipt".
	RefDocType string `gorm:"size:20;not null" json:"ref_doc_type"`
	RefDocID   uint   `gorm:"not null" json:"ref_doc_id"`
	RefDocNo   string `gorm:"size:50" json:"ref_doc_no,omitempty"`

	// System-set, read-only per §5.13's header table - same convention as
	// every other document in this codebase (e.g.
	// accounting_models.SalesInvoiceContent.DocDate): set client-side from
	// the user's PC clock at "New", not derived server-side from time.Now().
	DocDate string `json:"doc_date,omitempty"`

	// Reference SO and quote follow from RefDocID, not stored redundantly
	// here - they are looked up through the reference document.
	ExpectedReturnedDate string `json:"expected_returned_date,omitempty"`
	TransactionType      string `gorm:"size:50" json:"transaction_type,omitempty"`
	ShipTo               string `gorm:"size:255" json:"ship_to,omitempty"`
	LocationGroup        string `gorm:"size:100" json:"location_group,omitempty"`
	LocationCode         string `gorm:"size:100" json:"location_code,omitempty"`

	// Salesperson, Currency, and SalesPeriod come FROM the reference
	// document - never from the logged-in user, since A/R (not the
	// original sales exec) creates this document (§5.13, §3.3).
	Salesperson string `gorm:"size:255" json:"salesperson,omitempty"`
	Currency    string `gorm:"size:10" json:"currency,omitempty"`
	SalesPeriod string `gorm:"size:20" json:"sales_period,omitempty"`

	// Total = Σ TOTAL PRICE across lines. Server-computed and read-only;
	// stored here as a cache for listing/reporting, not as the source of
	// truth - recompute from SalesReturnDetails on every read that matters.
	Total float64 `json:"total"`

	// Approval gate (§3.3, §5.13): nothing fires on save. Stock movement
	// and the auto-generated Purchase Return happen only once IsApproved
	// flips true. The approver's name MUST be displayed, same as the SO
	// (§3.4).
	IsApproved     bool   `json:"is_approved"`
	ApprovedByID   uint   `json:"approved_by_id,omitempty"`
	ApprovedByName string `gorm:"size:255" json:"approved_by_name,omitempty"`
	ApprovalDate   string `json:"approval_date,omitempty"`

	// Pre-fills the eventual Credit Memo's required REASON CODE (§5.13).
	// Optional here - the memo itself still requires one at save.
	CmReasonCode string `gorm:"size:50" json:"cm_reason_code,omitempty"`

	// Read-only. Populated once GENERATE CREDIT MEMO (§5.13) produces a
	// CM# - blank where no credit was granted. Nothing in this model
	// creates that memo; this is just where its reference lands.
	RefCmID uint   `json:"ref_cm_id,omitempty"`
	RefCmNo string `gorm:"size:50" json:"ref_cm_no,omitempty"`

	HeaderRemarks string `gorm:"type:text" json:"header_remarks,omitempty"`
	Description   string `gorm:"type:text" json:"description,omitempty"`

	SalesReturnDetails []SalesReturnDetails `gorm:"foreignKey:SalesReturnID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sales_return_details,omitempty"`
}

type SalesReturn struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	DocNo int  `gorm:"not null;uniqueIndex:idx_tbl_trans_sales_return_doc_no" json:"doc_no"`
	SalesReturnContent
}

func (SalesReturn) TableName() string {
	return "tbl_trans_sales_return"
}

type SalesReturnAt struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesReturnContent
	At
}

func (SalesReturnAt) TableName() string {
	return "z_tbl_trans_sales_return_at"
}

// Request/response DTO shapes - same Body (create request) / Get (list
// response) split used by every other document in this codebase (e.g.
// accounting_models.SalesInvoiceBody/SalesInvoiceGet).
type SalesReturnBody struct {
	SalesReturn        SalesReturn          `json:"sales_return"`
	SalesReturnDetails []SalesReturnDetails `json:"sales_return_details"`
}

type SalesReturnGet struct {
	SalesReturn        []SalesReturn        `json:"sales_return"`
	SalesReturnDetails []SalesReturnDetails `json:"sales_return_details"`
}
