package item_stock_services

import (
	"strconv"
	"strings"
	"time"

	"github.com/pierceperado/smpc/models/inventory_models"
)

// Reservation windows and decisions - spec 10.4.2 and 10.4.5. A reservation's limit is its
// quotation's VALID UNTIL, and the two always move together.

// quoteDateTimeFormat is how tbl_trans_sales_quotation stores date and valid_until.
const quoteDateTimeFormat = "2006-01-02 15:04:05"

// defaultQuoteValidityDays is spec 5.1's default VALIDITY (DAYS).
const defaultQuoteValidityDays = 30

// quoteValidityDays reads a quotation's VALIDITY (DAYS), falling back to the spec default
// when it is blank or not a positive whole number.
func quoteValidityDays(raw string) int {
	days, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || days <= 0 {
		return defaultQuoteValidityDays
	}
	return days
}

// reservationWindowEnd is a fresh window: the given moment plus the quote's validity.
func reservationWindowEnd(from time.Time, validityDays string) time.Time {
	return from.AddDate(0, 0, quoteValidityDays(validityDays))
}

// parseQuoteDateTime reads the date formats found on quotations and in reservation
// requests. Formats without a zone are read as local time - they come off the user's PC
// clock, which CLAUDE.md makes the system's "now".
func parseQuoteDateTime(value string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &t
	}
	for _, layout := range []string{quoteDateTimeFormat, "2006-01-02T15:04:05", "2006-01-02", "01/02/2006 3:04:05 PM", "01/02/2006"} {
		if t, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return &t
		}
	}
	return nil
}

// sameMoment compares two quotation date strings as times, to the minute, so a formatting
// difference between what a client sends and what the table holds is not taken for an edit.
func sameMoment(a, b string) bool {
	ta, tb := parseQuoteDateTime(a), parseQuoteDateTime(b)
	if ta == nil || tb == nil {
		return strings.TrimSpace(a) == strings.TrimSpace(b)
	}
	return ta.Truncate(time.Minute).Equal(tb.Truncate(time.Minute))
}

// SameQuoteMoment is sameMoment for the quotation save paths in sales_services.
func SameQuoteMoment(a, b string) bool {
	return sameMoment(a, b)
}

// canDecide is the Reservations submodule's approve/decline rule (10.4.2). The Warehouse
// Manager "may decline or reverse any reservation", so any reservation may be approved or
// declined except into the state it is already in.
func canDecide(current, target string) bool {
	if target != inventory_models.ReservationStatusApproved && target != inventory_models.ReservationStatusRejected {
		return false
	}
	return current != target
}
