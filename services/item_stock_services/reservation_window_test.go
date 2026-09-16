package item_stock_services

import (
	"testing"
	"time"

	"github.com/pierceperado/smpc/models/inventory_models"
)

func TestQuoteValidityDaysFallsBackToTheSpecDefault(t *testing.T) {
	cases := map[string]int{"30": 30, " 45 ": 45, "": 30, "0": 30, "-5": 30, "thirty": 30}
	for raw, want := range cases {
		if got := quoteValidityDays(raw); got != want {
			t.Errorf("quoteValidityDays(%q) = %d, want %d", raw, got, want)
		}
	}
}

// Spec 10.4.5: keeping a reservation on hold, or changing its quantity, starts a fresh
// window from that moment.
func TestReservationWindowEndStartsFromTheGivenMoment(t *testing.T) {
	from := time.Date(2026, 9, 15, 10, 30, 0, 0, time.Local)

	if got, want := reservationWindowEnd(from, "30"), time.Date(2026, 10, 15, 10, 30, 0, 0, time.Local); !got.Equal(want) {
		t.Errorf("30-day window = %v, want %v", got, want)
	}
	if got, want := reservationWindowEnd(from, ""), from.AddDate(0, 0, 30); !got.Equal(want) {
		t.Errorf("blank validity = %v, want the 30-day default %v", got, want)
	}
}

func TestParseQuoteDateTimeReadsTheStoredFormats(t *testing.T) {
	cases := []struct {
		raw  string
		want time.Time
	}{
		{"2026-10-07 09:00:05", time.Date(2026, 10, 7, 9, 0, 5, 0, time.Local)},
		{"2026-10-07", time.Date(2026, 10, 7, 0, 0, 0, 0, time.Local)},
		{"10/07/2026 9:00:05 AM", time.Date(2026, 10, 7, 9, 0, 5, 0, time.Local)},
		{"2026-10-02T00:00:00+00:00", time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		got := parseQuoteDateTime(c.raw)
		if got == nil || !got.Equal(c.want) {
			t.Errorf("parseQuoteDateTime(%q) = %v, want %v", c.raw, got, c.want)
		}
	}

	if parseQuoteDateTime("") != nil || parseQuoteDateTime("soon") != nil {
		t.Error("blank or unreadable dates must come back nil")
	}
}

func TestSameMomentIgnoresFormatDifferences(t *testing.T) {
	if !sameMoment("2026-10-07 09:00:05", "10/07/2026 9:00:59 AM") {
		t.Error("the same minute in two formats should match")
	}
	if sameMoment("2026-10-07 09:00:05", "2026-10-08 09:00:05") {
		t.Error("different days must not match")
	}
}

// Spec 10.4.2: the Warehouse Manager may approve, decline, and reverse any reservation.
func TestCanDecideLetsTheWarehouseManagerReverseAnyDecision(t *testing.T) {
	pending, approved, rejected := inventory_models.ReservationStatusPending, inventory_models.ReservationStatusApproved, inventory_models.ReservationStatusRejected

	cases := []struct {
		from, to string
		want     bool
	}{
		{pending, approved, true},
		{pending, rejected, true},
		{approved, rejected, true}, // withdraw an approval
		{rejected, approved, true}, // re-tick a declined one
		{approved, approved, false},
		{rejected, rejected, false},
		{approved, pending, false}, // Pending is only ever a new request
	}

	for _, c := range cases {
		if got := canDecide(c.from, c.to); got != c.want {
			t.Errorf("canDecide(%s, %s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}
