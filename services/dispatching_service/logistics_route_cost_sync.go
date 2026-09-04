package dispatching_services

import (
	"errors"
	"strconv"
	"strings"

	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"gorm.io/gorm"
)

// User decision, 2026-09-03: there is now ONE delivery-cost table. The route's costs
// and the Delivery Receipt's DELIVERY COST block are the same rows in
// tbl_dispatching_delivery_receipt_costs - a route cost carries route_id, and also
// carries delivery_receipt_id when the route names a receipt. Nothing is copied
// between tables any more, so there is no sync that can silently fail to run.
//
// This prepares a schedule's route costs immediately before they are written:
// resolving COST TYPE both ways, computing TOTAL COST, and stamping the receipt the
// route names. RouteId itself is set by GORM when it writes the association.
func NormalizeRouteCosts(tx *gorm.DB, routes []dispatching_models.LogisticsRoute) error {
	for i := range routes {
		// A route need not have a Delivery Receipt at all - an office-to-office trip,
		// or a pick-up against a PO (§13.3). Those costs are stored against the route
		// alone, with delivery_receipt_id left at 0.
		receiptID, err := deliveryReceiptIDForDoc(tx, routes[i].DeliveryReceiptDoc)
		if err != nil {
			return err
		}

		for j := range routes[i].Costs {
			cost := &routes[i].Costs[j]

			// Incoming ids are from the previous save; the rows are being rewritten.
			cost.ID = 0
			cost.DeliveryReceiptID = receiptID

			if err := normalizeCostType(tx, cost); err != nil {
				return err
			}

			// TOTAL COST = AMOUNT × MULTIPLIER (§5.12). The route form posts neither
			// the product nor a total column, so it is computed on the way in.
			cost.TotalCost = cost.Amount * cost.Multiplier
		}
	}
	return nil
}

// NormalizeDeliveryReceiptCosts is the Delivery Receipt side of the same preparation.
// The DR posts COST TYPE as an id (its grid binds a combo to cost_type_id); without
// this the shared row would be written with an empty cost_type and the logistics route
// grid, which reads the name, would show a blank COST TYPE column.
func NormalizeDeliveryReceiptCosts(tx *gorm.DB, costs []dispatching_models.DeliveryReceiptCosts) error {
	for i := range costs {
		costs[i].ID = 0
		if err := normalizeCostType(tx, &costs[i]); err != nil {
			return err
		}
		costs[i].TotalCost = costs[i].Amount * costs[i].Multiplier
	}
	return nil
}

// DeleteRouteCostsForSchedule clears the cost rows belonging to a schedule's routes.
// UpdateLogisticsSchedule replaces routes wholesale (delete-then-reinsert), and the
// ON DELETE CASCADE that would clear these is an AutoMigrate-created FK which is not
// guaranteed to exist on a restored database - so the rows, and the receipt files
// hanging off them, are removed explicitly.
func DeleteRouteCostsForSchedule(tx *gorm.DB, scheduleID uint) error {
	var routeIDs []uint
	if err := tx.Model(&dispatching_models.LogisticsRoute{}).
		Where("schedule_id = ?", scheduleID).
		Pluck("id", &routeIDs).Error; err != nil {
		return err
	}
	if len(routeIDs) == 0 {
		return nil
	}

	var costIDs []uint
	if err := tx.Model(&dispatching_models.DeliveryReceiptCosts{}).
		Where("route_id IN ?", routeIDs).
		Pluck("id", &costIDs).Error; err != nil {
		return err
	}
	if len(costIDs) > 0 {
		if err := tx.Where("delivery_receipt_cost_id IN ?", costIDs).
			Delete(&dispatching_models.ReceiptFile{}).Error; err != nil {
			return err
		}
	}

	return tx.Where("route_id IN ?", routeIDs).
		Delete(&dispatching_models.DeliveryReceiptCosts{}).Error
}

// The route stores the Delivery Receipt as its typed document number, the same
// resolution RecomputeSoItemStatusForDeliveryReceiptDoc uses. A blank or unmatched
// value is not an error: the dispatcher may schedule the route before raising the
// receipt, and the next save after it exists will attach the costs.
func deliveryReceiptIDForDoc(tx *gorm.DB, doc string) (uint, error) {
	docNo, err := strconv.Atoi(strings.TrimSpace(doc))
	if err != nil || docNo == 0 {
		return 0, nil
	}

	var receipt dispatching_models.DeliveryReceipt
	if err := tx.Where("doc_no = ?", docNo).First(&receipt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return receipt.ID, nil
}

// COST TYPE is held as both an id and a name on the shared row: the Delivery Receipt
// binds a combo to cost_type_id, the route form posts and displays the plain name.
// Whichever one arrives, this fills in the other so both screens render.
//
// NOTE (§17.5): Cost Type is specified as hard-coded - Labor, Vehicle, Fuel, Toll
// Gate, Insurance, Penalty, Others - and explicitly "not configurable in Setup".
// Creating a missing row below keeps "+ADD COST" working rather than dropping the
// cost silently, but it leans into the Setup-backed model the spec rules out.
// Reported separately; resolve it there, not here.
func normalizeCostType(tx *gorm.DB, cost *dispatching_models.DeliveryReceiptCosts) error {
	name := strings.TrimSpace(cost.CostType)

	// Only an id arrived (a row that came from the Delivery Receipt side) - look up
	// its name so the route grid has something to show.
	if name == "" {
		if cost.CostTypeID == 0 {
			return nil
		}
		var costType dispatching_models.CalendarCostTypeModel
		if err := tx.First(&costType, cost.CostTypeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		cost.CostType = costType.Name
		return nil
	}

	// Matching by name, case-insensitively, is how DeliveryReceiptUC already
	// recognises the seven standard types against whatever casing Setup holds.
	var costType dispatching_models.CalendarCostTypeModel
	err := tx.Where("LOWER(name) = ?", strings.ToLower(name)).First(&costType).Error
	if err == nil {
		cost.CostTypeID = costType.ID
		cost.CostType = costType.Name
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	created := dispatching_models.CalendarCostTypeModel{
		CalendarCostTypeContent: dispatching_models.CalendarCostTypeContent{Name: name},
	}
	if err := tx.Create(&created).Error; err != nil {
		return err
	}
	cost.CostTypeID = created.ID
	cost.CostType = created.Name
	return nil
}
