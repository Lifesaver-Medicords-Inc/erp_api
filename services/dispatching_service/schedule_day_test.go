package dispatching_services

import (
	"testing"
	"time"
)

// The two shapes actually present in tbl_dispatching_logistics_calendar_schedule
// (ISO from the calendar UI, MM/dd/yyyy copied through from the Delivery
// Receipt) must resolve to the same day, or schedule grouping silently fails and
// every DR gets its own calendar entry again.
func TestParseScheduleDayAcceptsBothStoredFormats(t *testing.T) {
	want := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)

	cases := []string{
		"2026-09-04T00:00:00",
		"2026-09-04",
		"2026-09-04 00:00:00",
		"09/04/2026",
		"9/4/2026",
		"  09/04/2026  ",
	}

	for _, raw := range cases {
		got, ok := parseScheduleDay(raw)
		if !ok {
			t.Errorf("parseScheduleDay(%q) failed to parse", raw)
			continue
		}
		if !got.Equal(want) {
			t.Errorf("parseScheduleDay(%q) = %s, want %s", raw, got, want)
		}
	}
}

// MM/dd, not dd/MM: the client is en-PH, whose short-date pattern is M/d/yyyy.
// Reading "09/04/2026" as 9 April would file the schedule five months early.
func TestParseScheduleDayIsMonthFirst(t *testing.T) {
	got, ok := parseScheduleDay("09/04/2026")
	if !ok {
		t.Fatal("failed to parse")
	}
	if got.Month() != time.September || got.Day() != 4 {
		t.Errorf("got %s, want September 4", got.Format("January 2"))
	}
}

func TestParseScheduleDayRejectsUnusable(t *testing.T) {
	for _, raw := range []string{"", "   ", "not a date"} {
		if _, ok := parseScheduleDay(raw); ok {
			t.Errorf("parseScheduleDay(%q) reported success", raw)
		}
	}
}
