package item_stock_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"gorm.io/gorm"
)

// Negative stock on vehicle zones - spec 10.5, CLAUDE.md invariant 4.
//
// Item Release is the only document allowed to take a bin below zero, and only when that
// bin is a vehicle zone: "the rider picks up our purchased stock and drives straight to
// the client". The Receiving Report the rider then makes against the vehicle brings it
// back up. Item Request, Pick Activity, Stock Transfer and manual adjustment still refuse
// to take more than a bin holds.
//
// A stock row belongs to a vehicle when its warehouse and bin text are that vehicle zone's
// warehouse and location code (4.4.2: a vehicle zone's location code is its name alone).
// tbl_inv_item_stocks carries no zone id, so that is the only link available.

// mayGoBelowZero is the whole permission rule, kept apart so it can be tested on its own.
func mayGoBelowZero(allowVehicleNegative, isVehicleBin bool) bool {
	return allowVehicleNegative && isVehicleBin
}

// qtyBelowZero is how much of a deduction lands below zero: none while the bin still
// covers it, and all of it once the bin is already empty or negative.
func qtyBelowZero(before, deducted int) int {
	if deducted <= 0 {
		return 0
	}
	covered := before
	if covered < 0 {
		covered = 0
	}
	if deducted <= covered {
		return 0
	}
	return deducted - covered
}

// qtyOwedBack is how much of an incoming quantity only makes up for a bin that stood below
// zero - the units a vehicle already handed over before the stock was received.
func qtyOwedBack(before, incoming int) int {
	if before >= 0 || incoming <= 0 {
		return 0
	}
	if incoming < -before {
		return incoming
	}
	return -before
}

// costWithShortfall prices the below-zero units of a deduction at lastCost. Those units
// have no lot to draw from yet - the stock is still on its way in on a Receiving Report -
// and ConsumeLotsFIFO has already counted them at zero, so their share is added on top.
// Left unchanged when there is no known cost to use.
func costWithShortfall(cost *inventory_models.LotInfo, deducted, short int, lastCost float64) *inventory_models.LotInfo {
	if deducted <= 0 || short <= 0 || lastCost <= 0 {
		return cost
	}
	share := lastCost * float64(short) / float64(deducted)
	if cost == nil {
		return &inventory_models.LotInfo{UnitCost: share}
	}
	adjusted := *cost
	adjusted.UnitCost += share
	return &adjusted
}

// isVehicleBin reports whether a stock row's bin is a vehicle zone.
func (s *ItemStockService) isVehicleBin(tx *gorm.DB, warehouseId uint, binLocation string) (bool, error) {
	var count int64
	if err := tx.Model(&models.WarehouseArea{}).
		Where("vehicle_id <> 0 AND warehouse_name_id = ? AND location_code = ?", warehouseId, binLocation).
		Count(&count).Error; err != nil {
		return false, errors.New("failed checking whether the bin is a vehicle zone")
	}
	return count > 0, nil
}

// lastLotUnitCost is the unit cost of the item's most recently received lot in any bin,
// or 0 when it has never had one.
func (s *ItemStockService) lastLotUnitCost(tx *gorm.DB, itemId uint) (float64, error) {
	var lots []inventory_models.StockLot
	if err := tx.Where("item_id = ?", itemId).Order("id desc").Limit(1).Find(&lots).Error; err != nil {
		return 0, err
	}
	if len(lots) == 0 {
		return 0, nil
	}
	return lots[0].UnitCost, nil
}

// EnsureVehicleZoneStockRow returns the item's stock row on a vehicle zone, creating it at
// zero when the vehicle has never held the item. Item Release's picker offers every vehicle
// zone whether or not it holds stock (5.10), so the row a release deducts from may not
// exist yet. Refuses a zone that is not a vehicle.
func (s *ItemStockService) EnsureVehicleZoneStockRow(tx *gorm.DB, itemId, warehouseAreaId uint, uom string, atBody *inventory_models.ItemStocksAt, at models.At) (*inventory_models.ItemStocks, error) {
	var zones []models.WarehouseArea
	if err := tx.Where("id = ?", warehouseAreaId).Limit(1).Find(&zones).Error; err != nil {
		return nil, errors.New("failed reading the chosen zone")
	}
	if len(zones) == 0 {
		return nil, errors.New("the chosen vehicle zone no longer exists")
	}
	zone := zones[0]
	if zone.VehicleId == 0 {
		return nil, fmt.Errorf("zone %s is not a vehicle - only a vehicle zone can be released from without stock", zone.LocationCode)
	}

	var existing []inventory_models.ItemStocks
	if err := tx.Where("item_id = ? AND warehouse_id = ? AND bin_location = ?", itemId, zone.WarehouseNameId, zone.LocationCode).
		Limit(1).Find(&existing).Error; err != nil {
		return nil, errors.New("failed reading the vehicle's stock row")
	}
	if len(existing) > 0 {
		return &existing[0], nil
	}

	zero := 0
	body := &inventory_models.ItemStocks{
		ItemStocksContent: inventory_models.ItemStocksContent{
			ItemId:      itemId,
			StockQty:    &zero,
			StockUom:    uom,
			WarehouseId: zone.WarehouseNameId,
			BinLocation: zone.LocationCode,
		},
	}
	return s.UpsertStockWithTx(tx, body, atBody, at, nil)
}

// settleVehicleShortfall draws a newly received lot down by the units a vehicle already
// handed over while it stood below zero, booking the draw to the Item Release lines that
// took it there. Without this the lot would count every received unit as still on hand -
// overstating inventory value and never recording those units' cost of sales. Only what
// the releases still owe is drawn; anything beyond that stays in the lot.
func (s *ItemStockService) settleVehicleShortfall(tx *gorm.DB, stock *inventory_models.ItemStocks, owed int) error {
	if owed <= 0 {
		return nil
	}

	type openRelease struct {
		ID          uint
		SelectedQty int
		Consumed    int
	}
	var open []openRelease
	if err := tx.Raw(`
		SELECT l.id, l.selected_qty,
		       ISNULL((SELECT SUM(c.qty_consumed) FROM tbl_inv_stock_lot_consumptions c
		               WHERE c.ref_type = 'item_release' AND c.ref_id = l.id), 0) AS consumed
		FROM tbl_inv_item_release_locations l
		WHERE l.bin_id = ?
		ORDER BY l.id`, stock.ID).Scan(&open).Error; err != nil {
		return fmt.Errorf("failed reading the releases this vehicle owes: %w", err)
	}

	for _, r := range open {
		if owed <= 0 {
			break
		}
		take := r.SelectedQty - r.Consumed
		if take <= 0 {
			continue
		}
		if take > owed {
			take = owed
		}
		if _, err := s.ConsumeLotsFIFO(tx, stock.ItemId, stock.WarehouseId, stock.BinLocation, take, "item_release", r.ID); err != nil {
			return err
		}
		owed -= take
	}
	return nil
}
