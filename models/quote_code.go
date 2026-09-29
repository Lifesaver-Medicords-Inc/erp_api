package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// A quotation's full code: its number, version and sub-version as one string.
//
//	Q#2026-0001.1.0    a new quote (version 1, sub-version 0)
//	Q#2026-0001.2.0    New Version
//	Q#2026-0001.2.1    the first Edit session on version 2
//	FQ#2026-0001.2.1   that sub-version, finalized - it keeps all three parts
//
// (User decision 2026-09-25, from the SMPC versioning diagram; the year segment and the yearly
// restart are the spec's, §2.) The three parts are stored separately - document_no holds
// "Q#2026-0001", version_no and sub_version_no their numbers - and this is the ONE place they
// are put together, so every screen, list and print shows the same thing.
func FormatQuoteCode(documentNo, versionNo, subVersionNo string) string {
	doc := strings.TrimSpace(documentNo)
	if doc == "" {
		return ""
	}
	v := strings.TrimSpace(versionNo)
	if v == "" {
		v = "1"
	}
	s := strings.TrimSpace(subVersionNo)
	if s == "" {
		s = "0"
	}
	return fmt.Sprintf("%s.%s.%s", doc, v, s)
}

var quoteYearPattern = regexp.MustCompile(`(19|20)\d{2}`)

// QuoteYear is the year a quotation was raised, read from its DATE - the spec's rule for the
// year segment (§2: "the year the quotation was raised - not a fiscal year"). Falls back to
// the current year only when the date is missing or unreadable.
func QuoteYear(date string) int {
	d := strings.TrimSpace(date)
	for _, layout := range []string{
		"2006-01-02 15:04:05", "2006-01-02T15:04:05Z07:00", "2006-01-02",
		"01/02/2006 3:04:05 PM", "1/2/2006 3:04:05 PM", "01/02/2006", "1/2/2006",
	} {
		if t, err := time.Parse(layout, d); err == nil {
			return t.Year()
		}
	}
	if m := quoteYearPattern.FindString(d); m != "" {
		var y int
		if _, err := fmt.Sscanf(m, "%d", &y); err == nil {
			return y
		}
	}
	return time.Now().Year()
}

// AfterFind fills QuoteCode on every read, so every API response that carries a quotation -
// the list, the latest-versions view (which embeds this type), a search, a detail - hands the
// client the finished code instead of each screen assembling it.
func (q *SalesQuotation) AfterFind(tx *gorm.DB) error {
	q.QuoteCode = FormatQuoteCode(q.DocumentNo, q.VersionNo, q.SubVersionNo)
	return nil
}
