package inventory_models

import "time"

// StockReservation holds specific units against a specific quotation (spec 10.4). It never
// touches tbl_inv_item_stocks, tbl_inv_stock_lots or the trigger-driven ledger, because no
// units move: availability is computed as physical stock minus approved reservations (see
// GetAvailableStock).
//
// Raised from a quotation's RESERVE checkbox - Quick Quote lines as source_type
// "sales_quotation", Project Quote items as "sales_project_item". ExpiresAt is the quote's
// VALID UNTIL (10.4.5) and moves whenever that does.
//
// Not released when a quotation turns into a real order - there's no single, clean
// "quotation becomes an order" hookup in this codebase today, so double-counting between a
// lingering reservation and the real downstream deduction is a known possibility until
// that path exists.
//
// Status (10.4.2, 10.4.3): "Pending" the instant sales ticks RESERVE, "Approved" once the
// Warehouse Manager - any Position holding RESERVATION_APPROVAL - approves it, "Rejected"
// when declined or withdrawn (the row stays). Only Approved takes units out of
// availability: spec 14.22 forbids deducting stock for a reservation the Warehouse Manager
// has not approved.
const (
	ReservationStatusPending  = "Pending"
	ReservationStatusApproved = "Approved"
	ReservationStatusRejected = "Rejected"
)

type StockReservation struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	ItemId      uint       `json:"item_id"`
	Qty         uint       `json:"qty"`
	SourceType  string     `json:"source_type"`  // "sales_quotation" or "sales_project_item"
	SourceId    uint       `json:"source_id"`    // the quotation line's own id
	QuotationId uint       `json:"quotation_id"` // SalesQuotation.ID
	ReservedAt  time.Time  `json:"reserved_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Status      string     `gorm:"default:Pending" json:"status"`
	ApprovedBy  *uint      `json:"approved_by"`
	ApprovedAt  *time.Time `json:"approved_at"`

	// Set by the sweep once ExpiresAt has passed (10.4.5). Until the Warehouse Manager or the
	// owning sales executive answers - keep on hold or let go - the reservation keeps its
	// status and whatever stock it holds. Nothing releases or extends it on its own (14.23).
	LimitReachedAt *time.Time `json:"limit_reached_at"`
}

func (StockReservation) TableName() string {
	return "tbl_inv_stock_reservations"
}

// AvailableStockView is physical stock minus approved reservations for one item, summed
// across every bin/warehouse that item sits in.
type AvailableStockView struct {
	ItemId    uint `json:"item_id"`
	Physical  int  `json:"physical"`
	Reserved  int  `json:"reserved"`
	Available int  `json:"available"`
}

// PendingReservationView is one row of the Reservations submodule (10.4.2) - a
// StockReservation joined with just enough context (item name/model, the quotation it came
// from) to review it without a separate lookup per row. The sales red box reads the same
// rows for its at-limit questions.
// CustomerName / ProjectName come off the parent SalesQuotation so the approver can
// tell at a glance who the stock is being promised to - customer_id on the quotation
// is a tbl_bpi.id, whose display name lives one hop away in tbl_bpi_general.branch_name
// (same shape as the GetBpiCustomer view). A quotation can have no project name (Quick
// Quote), in which case ProjectName comes back empty rather than null.
type PendingReservationView struct {
	ID             uint       `json:"id"`
	ItemId         uint       `json:"item_id"`
	ItemName       string     `json:"item_name"`
	ItemModel      string     `json:"item_model"`
	ItemCode       string     `json:"item_code"`
	Qty            uint       `json:"qty"`
	SourceType     string     `json:"source_type"`
	SourceId       uint       `json:"source_id"`
	QuotationId    uint       `json:"quotation_id"`
	DocumentNo     string     `json:"document_no"`
	CustomerName   string     `json:"customer_name"`
	ProjectName    string     `json:"project_name"`
	RequestedBy    string     `json:"requested_by"`
	ReservedAt     time.Time  `json:"reserved_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	Status         string     `json:"status"`
	LimitReachedAt *time.Time `json:"limit_reached_at"`
}
