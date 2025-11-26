package setup_services

import (
	// "errors"

	"errors"
	"fmt"
	"time"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateInventoryStock(tx *gorm.DB, body *models.InventoryStocks, at models.At) error {

	// Create a Inventory for each detail entry
	inventory := models.InventoryStocks{
		InventoryStocksContent: body.InventoryStocksContent,
	}

	if err := services.DbInsert(tx, &inventory); err != nil {
		return errors.New("failed creating inventory stocks")
	}

	// Audit trail for each history record
	atdataInventory := models.InventoryStocksAt{
		RefId:                  inventory.ID,
		InventoryStocksContent: inventory.InventoryStocksContent,
		At:                     at,
	}

	if err := services.DbInsert(tx, &atdataInventory); err != nil {
		return errors.New("failed creating inventory stocks at")
	}

	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()
	return nil
}

func UpdateInventoryStock(tx *gorm.DB, body *models.InventoryStocks, at models.At) error {

	// Build or update the inventory record
	inventory := models.InventoryStocks{
		InventoryStocksContent: body.InventoryStocksContent,
	}

	// If an existing record exists (Receiving Report Details Id + Receiving report Id combination), update it
	var existing models.InventoryStocks
	err := tx.Where("receiving_report_id = ? AND receiving_report_details_id = ?", body.ReceivingReportId, body.ReceivingReportDetailsId).First(&existing).Error
	if err == nil {
		inventory.ID = existing.ID // ensure update, not insert
		if err := services.DbUpdate(tx, &inventory, map[string]interface{}{"id": existing.ID}); err != nil {
			return errors.New("failed updating inventory stocks")
		}
	} else {
		return errors.New("failed fetching inventory stocks for update")
	}

	// Insert audit trail record for inventory update
	atdataInventory := models.InventoryStocksAt{
		RefId:                  inventory.ID,
		InventoryStocksContent: inventory.InventoryStocksContent,
		At:                     at,
	}

	if err := services.DbInsert(tx, &atdataInventory); err != nil {
		return errors.New("failed creating inventory stocks audit record")
	}

	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return nil
}

func UpdateInventoryStockPickActivity(tx *gorm.DB, body *models.InventoryStocks, at models.At) error {

	// Build or update the inventory record
	inventory := models.InventoryStocks{
		InventoryStocksContent: body.InventoryStocksContent,
	}

	// If an existing record exists (Bin Location + Warehouse Id, Item Id combination), update it
	var existing models.InventoryStocks
	err := tx.Where("pick_activity_id = ? AND pick_activity_details_id = ?", inventory.PickActivityId, inventory.PickActivityDetailsId).First(&existing).Error
	if err == nil {
		if err := services.DbUpdate(tx, &inventory, map[string]interface{}{"id": existing.ID}); err != nil {
			return errors.New("failed updating inventory stocks")
		}
	} else {
		return errors.New("failed fetching inventory stocks for update")
	}

	// Insert audit trail record for inventory update
	atdataInventory := models.InventoryStocksAt{
		RefId:                  inventory.ID,
		InventoryStocksContent: inventory.InventoryStocksContent,
		At:                     at,
	}

	if err := services.DbInsert(tx, &atdataInventory); err != nil {
		return errors.New("failed creating inventory stocks audit record")
	}

	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return nil
}

func UpdateInventoryStocksHistory(tx *gorm.DB, body *models.InventoryStocksHistory, at models.At) error {

	// Build or update the inventory record
	inventory := models.InventoryStocksHistory{
		InventoryStocksHistoryContent: body.InventoryStocksHistoryContent,
	}

	var existing models.InventoryStocks
	err := tx.Where("item_id = ? AND bin_location = ? AND warehouse_id = ?", body.ItemId, body.BinLocation, body.WarehouseId).First(&existing).Error

	var existingHistory models.InventoryStocksHistory
	var errh error

	if body.PickActivityId == 0 && body.PickActivityDetailsId == 0 {
		// If both are zero → DO THIS QUERY
		errh = tx.Where(`item_id = ? AND bin_location = ? AND warehouse_id = ? 
			AND item_request_id = ? AND item_request_details_id = ?`,
			body.ItemId, body.BinLocation, body.WarehouseId, body.ItemRequestId, body.ItemRequestDetailsId).
			First(&existingHistory).Error

	} else {
		// Else → DO SOMETHING ELSE (replace with your logic)
		errh = tx.Where(`item_id = ? AND bin_location = ? AND warehouse_id = ?
			AND pick_activity_id = ? AND pick_activity_details_id = ?`,
			body.ItemId, body.BinLocation, body.WarehouseId, body.PickActivityId, body.PickActivityDetailsId).
			First(&existingHistory).Error
	}

	// Set current date in MM/dd/yyyy format
	inventory.TransactionDate = time.Now().Format("01/02/2006")

	if errh == nil {
		inventory.ID = existingHistory.ID
		inventory.InventoryStockId = existingHistory.InventoryStockId

		if err := services.DbUpdate(tx, &inventory, map[string]interface{}{"id": inventory.ID}); err != nil {
			return errors.New("failed updating inventory stocks history")
		}
	} else {
		// Save updated inventory
		inventory.InventoryStockId = existing.ID

		if err := services.DbInsert(tx, &inventory); err != nil {
			return errors.New("failed updating inventory stock location history")
		}
	}

	var totalReqQty int64
	tx.Model(&models.InventoryStocksHistory{}).
		Where("item_id = ? AND bin_location = ? AND warehouse_id = ?", body.ItemId, body.BinLocation, body.WarehouseId).
		Select("COALESCE(SUM(req_qty),0)").
		Scan(&totalReqQty)

	// Update QtyOut in InventoryStocks
	if err == nil {
		if err := services.DbUpdate(tx, &existing, map[string]interface{}{"id": existing.ID}); err != nil {
			return errors.New("failed updating total qty_out in inventory stocks")
		}
	}

	// Insert new inventory stock At (audit trail)
	inventoryAt := models.InventoryStocksHistoryAt{
		RefId:                         inventory.ID,
		InventoryStocksHistoryContent: inventory.InventoryStocksHistoryContent,
		At:                            at,
	}

	if err := services.DbInsert(tx, &inventoryAt); err != nil {
		return errors.New("failed creating inventory stocks at")
	}

	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}

func DeleteInventoryStock(tx *gorm.DB, pickActivityId uint, receivingId uint, at models.At) error {

	// Determine filter field
	var filter map[string]interface{}
	if pickActivityId != 0 {
		filter = map[string]interface{}{"pick_activity_id": pickActivityId}
	} else if receivingId != 0 {
		filter = map[string]interface{}{"receiving_report_id": receivingId}
	} else {
		return errors.New("no valid ID provided: both pickActivityId and receivingId are zero")
	}

	// Delete inventory stocks
	if err := services.DbDelete(tx, &models.InventoryStocks{}, filter); err != nil {
		return errors.New("failed deleting inventory stock")
	}

	// Fetch deleted rows (Unscoped for audit)
	var deletedInventories []models.InventoryStocks
	if err := tx.Unscoped().Where(filter).Find(&deletedInventories).Error; err == nil {
		for _, history := range deletedInventories {
			atdataInventory := models.InventoryStocksAt{
				RefId:                  history.ID,
				InventoryStocksContent: history.InventoryStocksContent,
				At:                     at,
			}
			if err := services.DbInsert(tx, &atdataInventory); err != nil {
				return errors.New("failed creating inventory stocks audit record")
			}
		}
	}

	// Cache invalidation
	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}

func DeleteInventoryStocksIRHistory(tx *gorm.DB, body *models.InventoryStocksHistory, at models.At) error {
	// Delete all inventory histories linked to the Item Request
	if err := services.DbDelete(tx, &models.InventoryStocksHistory{}, map[string]interface{}{"item_request_id": body.ItemRequestId}); err != nil {
		return errors.New("failed deleting all inventory stocks history")
	}

	// Optionally fetch deleted history records (Unscoped for audit)
	var deletedHistories []models.InventoryStocksHistory
	if err := tx.Unscoped().Where("item_request_id = ?", body.ItemRequestId).Find(&deletedHistories).Error; err == nil {
		for _, history := range deletedHistories {
			atdataHistory := models.InventoryStocksHistoryAt{
				RefId:                         history.ID,
				InventoryStocksHistoryContent: history.InventoryStocksHistoryContent,
				At:                            at,
			}
			if err := services.DbInsert(tx, &atdataHistory); err != nil {
				return errors.New("failed creating inventory stock history audit record")
			}
		}
	}

	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}

func DeleteInventoryStocksPAHistory(tx *gorm.DB, body *models.InventoryStocksHistory, at models.At) error {
	// Delete all inventory histories linked to the Pick Activity
	if err := services.DbDelete(tx, &models.InventoryStocksHistory{}, map[string]interface{}{"pick_activity_id": body.PickActivityId}); err != nil {
		return errors.New("failed deleting all inventory stocks history")
	}

	// Optionally fetch deleted history records (Unscoped for audit)
	var deletedHistories []models.InventoryStocksHistory
	if err := tx.Unscoped().Where("pick_activity_id = ?", body.PickActivityId).Find(&deletedHistories).Error; err == nil {
		for _, history := range deletedHistories {
			atdataHistory := models.InventoryStocksHistoryAt{
				RefId:                         history.ID,
				InventoryStocksHistoryContent: history.InventoryStocksHistoryContent,
				At:                            at,
			}
			if err := services.DbInsert(tx, &atdataHistory); err != nil {
				return errors.New("failed creating inventory stock history audit record")
			}
		}
	}

	if err := services.InvalidateCacheByModel(models.AllBinLocationView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	InvalidateItemCaches()

	return nil
}
