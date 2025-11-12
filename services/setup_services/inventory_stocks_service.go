package setup_services

import (
	// "errors"

	"errors"

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
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert new if not found
		if err := services.DbInsert(tx, &inventory); err != nil {
			return errors.New("failed creating new inventory stocks")
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
