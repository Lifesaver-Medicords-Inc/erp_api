package dispatching_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
)

// PickLocation is one row of Item Release's Actual Pick Qty modal.
type PickLocation struct {
	BinId           uint   `json:"bin_id"`
	BinLocation     string `json:"bin_location"`
	WarehouseId     uint   `json:"warehouse_id"`
	ItemId          uint   `json:"item_id"`
	StockQty        int    `json:"stock_qty"`
	StockUom        string `json:"stock_uom"`
	WarehouseAreaId uint   `json:"warehouse_area_id"`
	IsVehicle       bool   `json:"is_vehicle"`
}

// GetPickLocations lists where an Item Release line may take stock from: every bin holding
// the item, plus every vehicle zone whether it holds the item or not. Spec 5.10: the modal
// "includes vehicles by default, with or without stock in other bins", because a release
// may go below zero only on a vehicle (10.5). A vehicle with no stock row for the item comes
// back as bin_id 0 with its warehouse_area_id, and saving the release creates the row -
// nothing is written here.
//
// A vehicle row already below zero is inactive (SetActiveStatus) but still listed, so the
// same vehicle can keep carrying the item.
func (s *ItemReleaseService) GetPickLocations(itemId uint) ([]PickLocation, int, error) {
	rows := []PickLocation{}

	query := `
		SELECT s.id AS bin_id, s.bin_location, s.warehouse_id, s.item_id, s.stock_qty, s.stock_uom,
		       ISNULL(v.id, 0) AS warehouse_area_id,
		       CAST(CASE WHEN v.id IS NULL THEN 0 ELSE 1 END AS bit) AS is_vehicle
		FROM tbl_inv_item_stocks s
		OUTER APPLY (
			SELECT TOP 1 a.id FROM tbl_inv_warehouse_area a
			WHERE a.vehicle_id <> 0 AND a.warehouse_name_id = s.warehouse_id AND a.location_code = s.bin_location
		) v
		WHERE s.item_id = @item AND (s.is_active = 1 OR v.id IS NOT NULL)

		UNION ALL

		SELECT 0, a.location_code, a.warehouse_name_id, @item, 0,
		       ISNULL(
		           (SELECT TOP 1 s2.stock_uom FROM tbl_inv_item_stocks s2
		            WHERE s2.item_id = @item AND ISNULL(s2.stock_uom, '') <> ''
		            GROUP BY s2.stock_uom ORDER BY COUNT(*) DESC),
		           (SELECT u.name FROM tbl_setup_item i
		            JOIN tbl_setup_item_unit_measurement u ON u.id = i.unit_of_measure_id
		            WHERE i.id = @item)),
		       a.id, CAST(1 AS bit)
		FROM tbl_inv_warehouse_area a
		WHERE a.vehicle_id <> 0
		  AND NOT EXISTS (
			SELECT 1 FROM tbl_inv_item_stocks s
			WHERE s.item_id = @item AND s.warehouse_id = a.warehouse_name_id AND s.bin_location = a.location_code
		  )

		ORDER BY is_vehicle, bin_location`

	if err := initializers.DB.Raw(query, map[string]interface{}{"item": itemId}).Scan(&rows).Error; err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed getting pick locations")
	}

	return rows, fiber.StatusOK, nil
}
