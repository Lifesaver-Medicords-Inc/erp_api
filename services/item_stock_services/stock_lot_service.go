package item_stock_services

import (
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// CreateStockLot records one purchase batch for an item+warehouse+bin so a later sale
// can draw from it FIFO. Called by UpsertStockWithTx whenever it's given non-nil lot
// info (currently only Receiving Report provides this - see receiving_report_service.go).
func (s *ItemStockService) CreateStockLot(tx *gorm.DB, itemId, warehouseId uint, binLocation string, qty int, lot *inventory_models.LotInfo) error {
	if qty <= 0 {
		return nil
	}

	newLot := &inventory_models.StockLot{
		ItemId:       itemId,
		WarehouseId:  warehouseId,
		BinLocation:  binLocation,
		UnitCost:     lot.UnitCost,
		SupplierId:   lot.SupplierId,
		Supplier:     lot.Supplier,
		PurchaseDate: lot.PurchaseDate,
		QtyReceived:  qty,
		QtyRemaining: qty,
		SourceType:   lot.SourceType,
		SourceId:     lot.SourceId,
	}

	return services.DbInsert(tx, newLot)
}

// ConsumeLotsFIFO draws qty units out of the oldest available lot(s) for an
// item+warehouse+bin, decrementing each lot's remaining qty as it goes, and records a
// StockLotConsumption row per lot touched so ReleaseLotsFIFO can undo it exactly later.
// refType/refId identify the specific deduction event - DeductStockWithTx passes its
// own source_type/source_id, which is precise enough except for the one edge case
// documented on StockLotConsumption (multiple lines for the same item+bin under one
// document would share a ref and get their consumptions blended together).
//
// If the lots on hand don't cover the full qty (stock that predates lot tracking, or a
// manual add with no lot), the shortfall is treated as zero-cost rather than blocking
// the sale - the returned unit cost will just be a blend that reads lower than reality
// for that portion. Returns (nil, nil) when there's no lot history at all to report.
func (s *ItemStockService) ConsumeLotsFIFO(tx *gorm.DB, itemId, warehouseId uint, binLocation string, qty int, refType string, refId uint) (*inventory_models.LotInfo, error) {
	if qty <= 0 {
		return nil, nil
	}

	var lots []inventory_models.StockLot
	if err := tx.Where(
		"item_id = ? AND warehouse_id = ? AND bin_location = ? AND qty_remaining > 0",
		itemId, warehouseId, binLocation,
	).Order("purchase_date asc, id asc").Find(&lots).Error; err != nil {
		return nil, err
	}

	remaining := qty
	var totalCost float64
	var firstSupplierId uint
	var firstSupplier, firstPurchaseDate string
	haveFirst := false

	for i := range lots {
		if remaining <= 0 {
			break
		}
		lot := &lots[i]

		take := lot.QtyRemaining
		if take > remaining {
			take = remaining
		}

		lot.QtyRemaining -= take
		if err := services.DbUpdate(tx, lot, map[string]interface{}{"id": lot.ID}); err != nil {
			return nil, err
		}

		consumption := &inventory_models.StockLotConsumption{
			LotId:       lot.ID,
			RefType:     refType,
			RefId:       refId,
			QtyConsumed: take,
		}
		if err := services.DbInsert(tx, consumption); err != nil {
			return nil, err
		}

		totalCost += float64(take) * lot.UnitCost
		if !haveFirst {
			firstSupplierId = lot.SupplierId
			firstSupplier = lot.Supplier
			firstPurchaseDate = lot.PurchaseDate
			haveFirst = true
		}

		remaining -= take
	}

	if !haveFirst {
		// No lots at all for this item+bin - nothing to report, sale still proceeds.
		return nil, nil
	}

	// remaining > 0 here means the lots on hand didn't cover the full qty - the
	// uncovered portion is folded in at zero cost (see doc comment above).
	return &inventory_models.LotInfo{
		UnitCost:     totalCost / float64(qty),
		SupplierId:   firstSupplierId,
		Supplier:     firstSupplier,
		PurchaseDate: firstPurchaseDate,
	}, nil
}

// ReleaseLotsFIFO reverses whatever ConsumeLotsFIFO drew down for the given
// (refType, refId), adding qty back onto the exact lot(s) it came from - most
// recently consumed first. Used by RestoreStockWithTx. Returns (nil, nil) if there's
// nothing on record for that ref (e.g. the original deduction predates lot tracking).
func (s *ItemStockService) ReleaseLotsFIFO(tx *gorm.DB, refType string, refId uint) (*inventory_models.LotInfo, error) {
	var consumptions []inventory_models.StockLotConsumption
	if err := tx.Where("ref_type = ? AND ref_id = ?", refType, refId).
		Order("id desc").
		Find(&consumptions).Error; err != nil {
		return nil, err
	}

	if len(consumptions) == 0 {
		return nil, nil
	}

	var totalCost float64
	var totalQty int
	var firstSupplierId uint
	var firstSupplier, firstPurchaseDate string
	haveFirst := false

	for _, c := range consumptions {
		var lot inventory_models.StockLot
		if err := tx.Where("id = ?", c.LotId).First(&lot).Error; err != nil {
			return nil, err
		}

		lot.QtyRemaining += c.QtyConsumed
		if err := services.DbUpdate(tx, &lot, map[string]interface{}{"id": lot.ID}); err != nil {
			return nil, err
		}

		totalCost += float64(c.QtyConsumed) * lot.UnitCost
		totalQty += c.QtyConsumed
		if !haveFirst {
			firstSupplierId = lot.SupplierId
			firstSupplier = lot.Supplier
			firstPurchaseDate = lot.PurchaseDate
			haveFirst = true
		}

		if err := services.DbDelete(tx, &inventory_models.StockLotConsumption{}, map[string]interface{}{"id": c.ID}); err != nil {
			return nil, err
		}
	}

	if totalQty == 0 {
		return nil, nil
	}

	return &inventory_models.LotInfo{
		UnitCost:     totalCost / float64(totalQty),
		SupplierId:   firstSupplierId,
		Supplier:     firstSupplier,
		PurchaseDate: firstPurchaseDate,
	}, nil
}

// GetCostOfSales sums the actual FIFO cost of everything consumed (sold/released)
// between periodStart and periodEnd inclusive (both "YYYY-MM-DD") - built for the
// Admin "Reports" Income Statement (approved plan, increment 2), which needs Cost
// of Sales but has no Inventory-side report of its own to call yet.
//
// This is a perpetual-inventory answer, not the periodic Beginning+Purchases-Ending
// formula the reference financial statements use (Note 11) - it reads directly off
// tbl_inv_stock_lot_consumptions, which already records exactly what was drawn from
// which lot and when. Deliberately not reproducing the periodic formula: it would
// need a "beginning inventory" snapshot this system has no mechanism to produce,
// while this table already has the real answer recorded as it happened.
//
// created_at is a genuine datetime column (unlike the accounting side's free-text
// posting_date), so this compares directly - no TRY_CONVERT gymnastics needed.
//
// Known, inherited limitation (see ConsumeLotsFIFO's own doc comment): stock that
// predates lot tracking, or was added without a lot (e.g. a manual stock adjustment),
// consumes at zero recorded cost. This understates Cost of Sales for exactly that
// portion - flagged here rather than silently trusted, not fixed by this function.
func (s *ItemStockService) GetCostOfSales(periodStart, periodEnd string) (float64, error) {
	var totalCost float64

	query := `
		SELECT ISNULL(SUM(
			CAST(c.qty_consumed AS DECIMAL(18,2)) * CAST(l.unit_cost AS DECIMAL(18,4))
		), 0)
		FROM tbl_inv_stock_lot_consumptions c
		JOIN tbl_inv_stock_lots l ON c.lot_id = l.id
		WHERE c.created_at >= ? AND c.created_at < DATEADD(day, 1, CAST(? AS date))
	`

	if err := initializers.DB.Raw(query, periodStart, periodEnd).Row().Scan(&totalCost); err != nil {
		return 0, err
	}

	return totalCost, nil
}

// GetInventoryValue is the current total FIFO cost of everything still on hand
// (SUM(qty_remaining * unit_cost) across every lot) - the Balance Sheet's Inventory
// line and the Income Statement's Ending Inventory figure both need this. Always
// "as of now": there's no lot-level history snapshot mechanism to ask for a past
// date's value, so an as-of-date parameter would be misleading rather than useful -
// not offered here.
func (s *ItemStockService) GetInventoryValue() (float64, error) {
	var totalValue float64

	query := `
		SELECT ISNULL(SUM(
			CAST(qty_remaining AS DECIMAL(18,2)) * CAST(unit_cost AS DECIMAL(18,4))
		), 0)
		FROM tbl_inv_stock_lots
	`

	if err := initializers.DB.Raw(query).Row().Scan(&totalValue); err != nil {
		return 0, err
	}

	return totalValue, nil
}
