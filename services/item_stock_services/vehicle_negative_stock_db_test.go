//go:build dbtest

package item_stock_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
)

// Spec 10.5's worked example run against the database in .env, inside one transaction that
// is always rolled back. Run by hand only:
//
//	go test -tags dbtest -run TestVehicleNegativeStockAgainstDatabase -v ./services/item_stock_services/
func TestVehicleNegativeStockAgainstDatabase(t *testing.T) {
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
	at := models.At{AtUserId: "0", AtUser: "dbtest"}

	var zone models.WarehouseArea
	if err := tx.Where("vehicle_id <> 0").First(&zone).Error; err != nil {
		t.Skip("no vehicle zone in this database")
	}
	var normal inventory_models.ItemStocks
	if err := tx.Where("is_active = 1 AND stock_qty > 0").First(&normal).Error; err != nil {
		t.Skip("no stocked bin in this database")
	}
	reload := func(id uint) int {
		var r inventory_models.ItemStocks
		if err := tx.Where("id = ?", id).First(&r).Error; err != nil {
			t.Fatal(err)
		}
		return *r.StockQty
	}

	// 1) The vehicle's stock row, created at zero if the vehicle never held the item.
	row, err := s.EnsureVehicleZoneStockRow(tx, normal.ItemId, zone.ID, normal.StockUom,
		&inventory_models.ItemStocksAt{SourceType: "item_release", Remarks: "dbtest"}, at)
	if err != nil {
		t.Fatalf("ensure vehicle row: %v", err)
	}
	start := *row.StockQty
	if start < 0 {
		t.Skip("the vehicle is already below zero for this item")
	}

	// 2) Item Release takes 3 more than the vehicle holds.
	release := start + 3
	loc := models.ItemReleaseLocations{ItemReleaseLocationsContent: models.ItemReleaseLocationsContent{BinId: row.ID, SelectedQty: release}}
	if err := services.DbInsert(tx, &loc); err != nil {
		t.Fatalf("insert release location: %v", err)
	}
	if _, err := s.DeductStockForItemReleaseWithTx(tx, &inventory_models.ItemStocks{ID: row.ID, ItemStocksContent: inventory_models.ItemStocksContent{StockQty: &release}},
		&inventory_models.ItemStocksAt{SourceType: "item_release", SourceId: loc.ID, Remarks: "dbtest"}, at); err != nil {
		t.Fatalf("Item Release on a vehicle should go below zero: %v", err)
	}
	if got := reload(row.ID); got != -3 {
		t.Fatalf("vehicle after release = %d, want -3", got)
	}

	// Anything but Item Release is still refused on the vehicle...
	one := 1
	if _, err := s.DeductStockWithTx(tx, &inventory_models.ItemStocks{ID: row.ID, ItemStocksContent: inventory_models.ItemStocksContent{StockQty: &one}},
		&inventory_models.ItemStocksAt{SourceType: "item_request", Remarks: "dbtest"}, at); err == nil {
		t.Fatal("a non-release deduction took the vehicle further below zero")
	}
	// ...and Item Release is still capped on an ordinary bin.
	over := *normal.StockQty + 1
	if _, err := s.DeductStockForItemReleaseWithTx(tx, &inventory_models.ItemStocks{ID: normal.ID, ItemStocksContent: inventory_models.ItemStocksContent{StockQty: &over}},
		&inventory_models.ItemStocksAt{SourceType: "item_release", Remarks: "dbtest"}, at); err == nil {
		t.Fatal("Item Release took an ordinary bin below zero")
	}

	// 3) The rider's Receiving Report of 5 against the vehicle.
	five := 5
	if _, err := s.UpsertStockWithTx(tx, &inventory_models.ItemStocks{ItemStocksContent: inventory_models.ItemStocksContent{
		ItemId: normal.ItemId, StockQty: &five, StockUom: normal.StockUom, WarehouseId: zone.WarehouseNameId, BinLocation: zone.LocationCode,
	}}, &inventory_models.ItemStocksAt{SourceType: "receiving_report", Remarks: "dbtest"}, at,
		&inventory_models.LotInfo{UnitCost: 100, SourceType: "receiving_report"}); err != nil {
		t.Fatalf("receiving report: %v", err)
	}
	if got := reload(row.ID); got != 2 {
		t.Fatalf("vehicle after receiving 5 = %d, want 2", got)
	}

	// Find + Limit, not First: First appends ORDER BY id, and SQL Server rejects id twice.
	var lots []inventory_models.StockLot
	if err := tx.Where("item_id = ? AND warehouse_id = ? AND bin_location = ?", normal.ItemId, zone.WarehouseNameId, zone.LocationCode).
		Order("id desc").Limit(1).Find(&lots).Error; err != nil || len(lots) == 0 {
		t.Fatalf("new lot: %v (found %d)", err, len(lots))
	}
	lot := lots[0]
	if lot.QtyReceived != 5 || lot.QtyRemaining != 2 {
		t.Fatalf("lot received/remaining = %d/%d, want 5/2", lot.QtyReceived, lot.QtyRemaining)
	}

	var consumed int
	if err := tx.Raw(`SELECT ISNULL(SUM(qty_consumed), 0) FROM tbl_inv_stock_lot_consumptions WHERE ref_type = 'item_release' AND ref_id = ? AND lot_id = ?`,
		loc.ID, lot.ID).Scan(&consumed).Error; err != nil {
		t.Fatal(err)
	}
	if consumed != 3 {
		t.Fatalf("units booked from the new lot to the release = %d, want 3", consumed)
	}

	t.Logf("vehicle %s: %d -> -3 after release, 2 after receiving 5; lot %d kept 2, 3 booked to release location %d (rolled back)",
		zone.LocationCode, start, lot.ID, loc.ID)
}
