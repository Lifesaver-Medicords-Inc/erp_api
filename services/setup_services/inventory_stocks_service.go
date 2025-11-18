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

	InvalidateItemCaches()
	return nil
}

func UpdateInventoryRRStock(tx *gorm.DB, body *models.InventoryStocks, at models.At) error {

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

	return nil
}

func CreateInventoryStocksHistory(tx *gorm.DB, body *models.InventoryStocksHistory, at models.At) error {

	fmt.Println("Qtys: ", body.ReqQty, body.StockQty)

	// Build or update the inventory record
	inventory := models.InventoryStocksHistory{
		InventoryStocksHistoryContent: body.InventoryStocksHistoryContent,
	}

	fmt.Println("Qtys: ", body.ReqQty, body.StockQty)

	var existing models.InventoryStocks
	err := tx.Where("item_id = ? AND bin_location = ?", body.ItemId, body.BinLocation).First(&existing).Error

	if err == nil {
		inventory.InventoryStockId = existing.ID
	} else {
		return errors.New("failed getting inventory stock location record")
	}

	// Set current date in MM/dd/yyyy format
	inventory.TransactionDate = time.Now().Format("01/02/2006")

	fmt.Println("Qtys: ", body.ReqQty, body.StockQty)

	// Save updated inventory
	if err := services.DbInsert(tx, &inventory); err != nil {
		return errors.New("failed updating inventory stock location history")
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

func UpdateInventoryStocksHistory(tx *gorm.DB, body *models.InventoryStocksHistory, at models.At) error {

	// Build or update the inventory record
	inventory := models.InventoryStocksHistory{
		InventoryStocksHistoryContent: body.InventoryStocksHistoryContent,
	}

	var existing models.InventoryStocks
	err := tx.Where("item_id = ? AND bin_location = ?", body.ItemId, body.BinLocation).First(&existing).Error

	if err == nil {
		inventory.InventoryStockId = existing.ID
	} else {
		return errors.New("failed getting inventory stock location record")
	}

	// Set current date in MM/dd/yyyy format
	inventory.TransactionDate = time.Now().Format("01/02/2006")

	// Save updated inventory
	if err := services.DbInsert(tx, &inventory); err != nil {
		return errors.New("failed updating inventory stock location history")
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

func DeleteInventoryStock(tx *gorm.DB, receivingId uint, at models.At) error {
	// Delete all inventory stocks linked to the Receiving Report
	if err := services.DbDelete(tx, &models.InventoryStocks{}, map[string]interface{}{"receiving_report_id": receivingId}); err != nil {
		return errors.New("failed deleting all inventory stock")
	}

	// Optionally fetch deleted inventory stock records (Unscoped for audit)
	var deletedInventories []models.InventoryStocks
	if err := tx.Unscoped().Where("receiving_report_id = ?", receivingId).Find(&deletedInventories).Error; err == nil {
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

	InvalidateItemCaches()

	return nil
}

func DeleteInventoryStocksHistory(tx *gorm.DB, body *models.InventoryStocksHistory, at models.At) error {
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

	InvalidateItemCaches()

	return nil
}
