//go:build dbtest

package item_stock_services

import (
	"os"
	"testing"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models/inventory_models"
)

// Spec 10.4.3-10.4.5 run against the database in .env, inside one transaction that is always
// rolled back. setReservationDecision and AnswerReservationLimit open transactions of their
// own, so approval is simulated here and the answers are exercised through the tx-level
// functions they call. Run by hand only:
//
//	go test -tags dbtest -run TestReservationsAgainstDatabase -v ./services/item_stock_services/
func TestReservationsAgainstDatabase(t *testing.T) {
	// Only once per process: the other DB test in this package moves to the module root too.
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()

	s := NewItemStockService()
	const source, sourceId = "dbtest", uint(990001)

	var quote struct {
		ID           uint
		ValidityDays string
		ValidUntil   string
	}
	if err := tx.Raw(`SELECT TOP 1 id, validity_days, valid_until FROM tbl_trans_sales_quotation WHERE ISNULL(valid_until, '') <> '' ORDER BY id DESC`).Scan(&quote).Error; err != nil || quote.ID == 0 {
		t.Skip("no quotation with a VALID UNTIL in this database")
	}
	var itemId uint
	if err := tx.Raw(`SELECT TOP 1 item_id FROM tbl_inv_item_stocks WHERE stock_qty > 0 ORDER BY id`).Scan(&itemId).Error; err != nil || itemId == 0 {
		t.Skip("no stocked item in this database")
	}

	available := func() int {
		n, err := s.getAvailableSnapshot(tx, itemId)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	reload := func() inventory_models.StockReservation {
		var rows []inventory_models.StockReservation
		if err := tx.Where("source_type = ? AND source_id = ?", source, sourceId).Limit(1).Find(&rows).Error; err != nil {
			t.Fatal(err)
		}
		if len(rows) == 0 {
			return inventory_models.StockReservation{}
		}
		return rows[0]
	}
	sameMinute := func(a, b time.Time) bool {
		d := a.Sub(b)
		return d < time.Minute && d > -time.Minute
	}
	before := available()

	// 1) A request moves no stock (14.22) and takes the quote's VALID UNTIL as its limit.
	if err := s.CreateStockReservation(tx, itemId, 2, source, sourceId, quote.ID, nil, "dbtest"); err != nil {
		t.Fatalf("create: %v", err)
	}
	r := reload()
	limit := parseQuoteDateTime(quote.ValidUntil)
	if r.Status != inventory_models.ReservationStatusPending || r.ExpiresAt == nil || limit == nil || !sameMinute(*r.ExpiresAt, *limit) {
		t.Fatalf("new request = %+v, want Pending expiring at the quote's VALID UNTIL %q", r, quote.ValidUntil)
	}
	if got := available(); got != before {
		t.Fatalf("a pending request changed availability: %d -> %d", before, got)
	}

	// 2) Approval takes the units out.
	if err := tx.Model(&inventory_models.StockReservation{}).Where("id = ?", r.ID).Update("status", inventory_models.ReservationStatusApproved).Error; err != nil {
		t.Fatal(err)
	}
	if got := available(); got != before-2 {
		t.Fatalf("approved availability = %d, want %d", got, before-2)
	}

	// 3) A quantity change is a new request with a fresh window, and the quote's VALID UNTIL
	// moves with it (14.25).
	changedAt := time.Now()
	if err := s.SyncReservationQty(tx, source, sourceId, 3); err != nil {
		t.Fatalf("sync qty: %v", err)
	}
	r = reload()
	wantEnd := reservationWindowEnd(changedAt, quote.ValidityDays)
	if r.Status != inventory_models.ReservationStatusPending || r.Qty != 3 || r.ExpiresAt == nil || !sameMinute(*r.ExpiresAt, wantEnd) {
		t.Fatalf("after qty change = %+v, want Pending qty 3 expiring about %v", r, wantEnd)
	}
	var validUntil string
	if err := tx.Raw(`SELECT valid_until FROM tbl_trans_sales_quotation WHERE id = ?`, quote.ID).Scan(&validUntil).Error; err != nil {
		t.Fatal(err)
	}
	if moved := parseQuoteDateTime(validUntil); moved == nil || !sameMinute(*moved, wantEnd) {
		t.Fatalf("quote VALID UNTIL = %q, want about %v", validUntil, wantEnd)
	}
	if got := available(); got != before {
		t.Fatalf("a changed request should hold nothing until approved again: %d, want %d", got, before)
	}

	// 4) At its limit the sweep flags it and releases nothing (14.23).
	if err := tx.Model(&inventory_models.StockReservation{}).Where("id = ?", r.ID).
		Updates(map[string]interface{}{"status": inventory_models.ReservationStatusApproved, "expires_at": time.Now().Add(-time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}
	if n, err := s.FlagReservationsAtLimit(tx); err != nil || n < 1 {
		t.Fatalf("flag at limit: n=%d err=%v", n, err)
	}
	r = reload()
	if r.ID == 0 || r.LimitReachedAt == nil || r.Status != inventory_models.ReservationStatusApproved {
		t.Fatalf("at limit = %+v, want the row kept, still Approved, flagged", r)
	}
	if got := available(); got != before-3 {
		t.Fatalf("at limit the units must stay held: %d, want %d", got, before-3)
	}

	// 5) Keep on hold: a fresh window, the question cleared, the stock still held.
	if _, err := s.restartQuoteWindow(tx, r.QuotationId, r.ID, time.Now()); err != nil {
		t.Fatalf("keep on hold: %v", err)
	}
	r = reload()
	if r.LimitReachedAt != nil || r.ExpiresAt == nil || !r.ExpiresAt.After(time.Now()) || r.Status != inventory_models.ReservationStatusApproved {
		t.Fatalf("after keep on hold = %+v, want unflagged, Approved, expiring in the future", r)
	}
	if got := available(); got != before-3 {
		t.Fatalf("kept on hold availability = %d, want %d", got, before-3)
	}

	// 6) Let go: the row goes and its units return.
	if err := s.deleteReservation(tx, &r, "dbtest", "dbtest let go"); err != nil {
		t.Fatalf("let go: %v", err)
	}
	if reload().ID != 0 {
		t.Fatal("let go left the reservation behind")
	}
	if got := available(); got != before {
		t.Fatalf("after let go availability = %d, want %d", got, before)
	}

	t.Logf("item %d on quote %d: available %d, pending %d, approved %d, changed %d, at limit %d, kept %d, let go %d (rolled back)",
		itemId, quote.ID, before, before, before-2, before, before-3, before-3, before)
}
