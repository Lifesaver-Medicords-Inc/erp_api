package inventory_models

import "time"

// StockReservation is a soft hold placed by a Sales Quotation line - it never touches
// tbl_inv_item_stocks, tbl_inv_stock_lots, or the ledger, because nothing physically
// moves when a quotation is created. It exists purely so "available to quote" can be
// computed as physical stock minus whatever's already promised elsewhere (see
// GetAvailableStock).
//
// Created immediately when a SalesQuotationQuick line is added (see
// quick_quotation_service.go), and released either when that line is deleted, or by
// the periodic sweep in ExpireStockReservations once ExpiresAt passes (copied from the
// parent SalesQuotation's ValidUntil at creation time - nil if that couldn't be
// parsed, meaning the sweep will never pick it up and it needs manual cleanup).
//
// Not released when a quotation turns into a real order - there's no single, clean
// "quotation becomes an order" hookup in this codebase today (see the conversation
// this was scoped in), so double-counting between a lingering reservation and the
// real downstream deduction is a known possibility until that path exists.
type StockReservation struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	ItemId      uint       `json:"item_id"`
	Qty         uint       `json:"qty"`
	SourceType  string     `json:"source_type"` // "sales_quotation" today
	SourceId    uint       `json:"source_id"`   // SalesQuotationQuick.ID
	QuotationId uint       `json:"quotation_id"` // SalesQuotation.ID (SalesQuotationQuick.BasedId)
	ReservedAt  time.Time  `json:"reserved_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

func (StockReservation) TableName() string {
	return "tbl_inv_stock_reservations"
}

// AvailableStockView is physical stock minus active reservations for one item, summed
// across every bin/warehouse that item sits in.
type AvailableStockView struct {
	ItemId    uint `json:"item_id"`
	Physical  int  `json:"physical"`
	Reserved  int  `json:"reserved"`
	Available int  `json:"available"`
}
