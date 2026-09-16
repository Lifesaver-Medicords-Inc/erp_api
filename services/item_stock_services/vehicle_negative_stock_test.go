package item_stock_services

import (
	"math"
	"testing"

	"github.com/pierceperado/smpc/models/inventory_models"
)

// Spec 14.15: nothing but Item Release, and nowhere but a vehicle zone, may go below zero.
func TestMayGoBelowZeroOnlyForItemReleaseOnAVehicle(t *testing.T) {
	cases := []struct {
		allowVehicleNegative, isVehicle, want bool
	}{
		{false, false, false},
		{false, true, false}, // Item Request, Pick Activity, Stock Transfer - even on a vehicle
		{true, false, false}, // Item Release on an ordinary bin
		{true, true, true},
	}

	for _, c := range cases {
		if got := mayGoBelowZero(c.allowVehicleNegative, c.isVehicle); got != c.want {
			t.Errorf("mayGoBelowZero(%v, %v) = %v, want %v", c.allowVehicleNegative, c.isVehicle, got, c.want)
		}
	}
}

// Spec 10.5's worked example: Truck 1 holds nothing and releases 3, so all 3 are below zero;
// the Receiving Report of 5 then owes 3 of them back and leaves 2 on hand.
func TestBelowZeroAndOwedBackFollowTheWorkedExample(t *testing.T) {
	below := []struct{ before, deducted, want int }{
		{0, 3, 3},
		{2, 2, 0},
		{2, 5, 3},
		{-3, 2, 2},
		{5, 0, 0},
	}
	for _, c := range below {
		if got := qtyBelowZero(c.before, c.deducted); got != c.want {
			t.Errorf("qtyBelowZero(%d, %d) = %d, want %d", c.before, c.deducted, got, c.want)
		}
	}

	owed := []struct{ before, incoming, want int }{
		{-3, 5, 3},
		{-3, 2, 2},
		{0, 5, 0},
		{4, 5, 0},
		{-3, 0, 0},
	}
	for _, c := range owed {
		if got := qtyOwedBack(c.before, c.incoming); got != c.want {
			t.Errorf("qtyOwedBack(%d, %d) = %d, want %d", c.before, c.incoming, got, c.want)
		}
	}
}

func TestCostWithShortfallPricesBelowZeroUnitsAtLastCost(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	// 5 released: 2 drawn from a lot at 100 (ConsumeLotsFIFO's blend over 5 is 40),
	// 3 below zero at the last known cost of 120.
	got := costWithShortfall(&inventory_models.LotInfo{UnitCost: 40, Supplier: "S"}, 5, 3, 120)
	if got == nil || !near(got.UnitCost, 112) || got.Supplier != "S" {
		t.Fatalf("mixed release: got %+v, want unit cost 112 keeping the lot's supplier", got)
	}

	// No lot at all: the whole release is below zero.
	if got := costWithShortfall(nil, 3, 3, 120); got == nil || !near(got.UnitCost, 120) {
		t.Fatalf("all below zero: got %+v, want unit cost 120", got)
	}

	// Nothing below zero, or no cost known: left as it was.
	original := &inventory_models.LotInfo{UnitCost: 50}
	if got := costWithShortfall(original, 4, 0, 120); got != original {
		t.Fatalf("nothing below zero should leave the cost alone, got %+v", got)
	}
	if got := costWithShortfall(nil, 3, 3, 0); got != nil {
		t.Fatalf("no known cost should leave the cost nil, got %+v", got)
	}
}
