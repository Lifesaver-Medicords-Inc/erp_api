package sales_services

import (
	"github.com/pierceperado/smpc/services/item_stock_services"
	"gorm.io/gorm"
)

// Quotation-side hooks into stock reservations (spec 10.4.5). A reservation's limit is its
// quote's VALID UNTIL, and a change to the reserved quantity is a new request with a fresh
// window - these keep the quotation save paths in step with that.

// projectItemReservationSource is the source_type Project Quotation items reserve under
// (the sales app's Quotation.cs / ItemSetUC).
const projectItemReservationSource = "sales_project_item"

// moveQuoteReservationLimits carries an edited VALID UNTIL onto the quote's reservations.
func moveQuoteReservationLimits(tx *gorm.DB, quotationId uint, validUntil string) error {
	return item_stock_services.NewItemStockService().MoveReservationLimits(tx, quotationId, validUntil)
}

// syncProjectItemReservation applies a project item's edited QTY to its reservation.
func syncProjectItemReservation(tx *gorm.DB, itemsId uint, qty uint) error {
	return item_stock_services.NewItemStockService().SyncReservationQty(tx, projectItemReservationSource, itemsId, qty)
}

// releaseProjectItemReservation drops the reservation of a project item being removed.
func releaseProjectItemReservation(tx *gorm.DB, itemsId uint, dbUser string) error {
	return item_stock_services.NewItemStockService().ReleaseStockReservation(tx, projectItemReservationSource, itemsId, dbUser)
}

// sameQuoteMoment compares two quotation date strings as moments, ignoring format.
func sameQuoteMoment(a, b string) bool {
	return item_stock_services.SameQuoteMoment(a, b)
}
