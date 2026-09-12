package models

// The charge record behind a Sales Order cancellation (spec 5.4, 8.15).
//
// Cancelling an approved SO is a two-stage act. Any sales executive may raise a
// cancellation on an order they own - they hold the client relationship and know
// the commercial terms - but it does not take effect until the Sales Manager or
// the CBDO approves it. Until then the SO sits at FOR REVIEW and NOTHING moves:
// stock stays reserved, every department's view is unchanged, and A/R sees
// nothing.
//
// One record per cancellation attempt. It is written when PROCEED is pressed and
// is what carries the negotiated percentages forward - the modal's defaults come
// from Company Setup (4.5.6) but are overridable, so the rate actually agreed
// with the client has to be stored here rather than re-read from setup later.
//
// The percentages are whole numbers (10 means 10%). Zero is a real, meaningful
// value: 8.15 says a 0 fee produces no invoice line, so a stored 0 means the fee
// was declined, never "unset".
type SalesOrderChargeContent struct {
	OrderID uint `json:"order_id"`
	// The order's document number at the time, so the record reads on its own.
	DocumentNo string `json:"document_no"`

	// Both percentages are always captured, with no checkbox anywhere - declining
	// a fee is entering 0 (5.4).
	RestockingFeePercent   float64 `json:"restocking_fee_percent"`
	CancellationFeePercent float64 `json:"cancellation_fee_percent"`

	// The computed figures, frozen here at PROCEED. They are stored rather than
	// recomputed on read because the fee base depends on which lines were
	// undelivered and on the VAT rate frozen on the SO (12.1) - both of which can
	// look different by the time anyone opens the record again.
	//
	// FeeBase is VAT-INCLUSIVE (8.15). Each fee therefore already contains VAT,
	// and the Sales Invoice line MUST back the VAT out rather than adding it -
	// adding VAT on top taxes the same amount twice.
	UndeliveredNet   float64 `json:"undelivered_net"`
	VatRatePercent   float64 `json:"vat_rate_percent"`
	FeeBase          float64 `json:"fee_base"`
	RestockingFee    float64 `json:"restocking_fee"`
	CancellationFee  float64 `json:"cancellation_fee"`
	TotalCharge      float64 `json:"total_charge"`

	// FOR REVIEW -> APPROVED | REJECTED. Rejection returns the SO to OPEN with no
	// charge record in force (5.4 step 5), so a rejected row is kept as history
	// rather than deleted.
	Status     string `json:"status"`
	Reason     string `json:"reason"`
	RaisedBy   string `json:"raised_by"`
	RaisedById uint   `json:"raised_by_id"`
	RaisedDate string `json:"raised_date"`
	// Sales Manager or CBDO only - the same pair that approves the order itself.
	ReviewedBy   string `json:"reviewed_by"`
	ReviewedById uint   `json:"reviewed_by_id"`
	ReviewedDate string `json:"reviewed_date"`
}

type SalesOrderCharge struct {
	ID uint `gorm:"primarykey" json:"id"`
	SalesOrderChargeContent
}

func (SalesOrderCharge) TableName() string {
	return "tbl_trans_sales_order_charge"
}

type SalesOrderChargeAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	SalesOrderChargeContent
	At
}

func (SalesOrderChargeAt) TableName() string {
	return "z_tbl_trans_sales_order_charge_at"
}
