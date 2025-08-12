package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// copy pasta ng rr deets, refactored
func GetReceivingReportInventory(conditions map[string]interface{}) ([]models.ReceivingReportInventory, int, error) {
	var Inventory []models.ReceivingReportInventory

	if err := services.DbGet(&Inventory, conditions); err != nil {
		return Inventory, fiber.StatusInternalServerError, errors.New("failed getting warehouse address")
	}

	return Inventory, 0, nil
}

func CreateReceivingReportInventory(tx *gorm.DB, parentId uint, child models.ReceivingReportInventory, at models.At) error {
	//pwedeng directa na since di nmn nag tthrow ng ID
	content := models.ReceivingReportInventoryContent{
		ReceivingReportId: parentId,
		ItemCode:          child.ItemCode,
		ItemDescription:   child.ItemDescription,
		OrderedQty:        child.OrderedQty,
		OrderedUom:        child.OrderedUom,
		SerialNumber:      child.SerialNumber,
		BinLocation:       child.BinLocation,
		RefId:             child.RefId,
	}
	Inventory := models.ReceivingReportInventory{
		ReceivingReportInventoryContent: content,
	}
	if err := services.DbInsert(tx, &Inventory); err != nil {
		return errors.New("failed creating warehouse area")
	}

	InventoryAt := models.ReceivingReportInventoryAt{
		RefId:                           Inventory.ID,
		ReceivingReportInventoryContent: content,
		At:                              at,
	}
	if err := services.DbInsert(tx, &InventoryAt); err != nil {
		return errors.New("failed creating warehouse area at")
	}

	return nil
}

func UpdateReceivingReportInventory(tx *gorm.DB, Inventory models.ReceivingReportInventory, at models.At, conditions map[string]interface{}) error {

	if err := services.DbUpdate(tx, &Inventory, conditions); err != nil {
		return errors.New("failed updating receiving report details")
	}

	InventoryAt := models.ReceivingReportInventoryAt{
		RefId:                           Inventory.ID,
		ReceivingReportInventoryContent: Inventory.ReceivingReportInventoryContent,
		At:                              at,
	}

	if err := services.DbInsert(tx, &InventoryAt); err != nil {
		return errors.New("failed creating receiving report details at")
	}

	return nil
}

func DeleteReceivingReportInventory(tx *gorm.DB, Inventory models.ReceivingReportInventory, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &Inventory, conditions); err != nil {
		return errors.New("failed deleting receiving report inventory")
	}

	Inventoryat := models.ReceivingReportInventoryAt{
		RefId:                           Inventory.ID,
		ReceivingReportInventoryContent: Inventory.ReceivingReportInventoryContent,
		At:                              at,
	}
	if err := services.DbInsert(tx, &Inventoryat); err != nil {
		return errors.New("failed creating receiving report inventory at")
	}

	return nil
}
