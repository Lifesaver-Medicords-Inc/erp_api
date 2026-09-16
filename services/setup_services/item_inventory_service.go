package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateItemInventory(tx *gorm.DB, basedId uint, itemInventory models.ItemInventory, at models.At) error {
	itemInventory.BasedId = basedId
	if err := services.DbInsert(tx, &itemInventory); err != nil {
		return errors.New("failed inserting item inventory")
	}
	itemInventoryAt := models.ItemInventoryAt{
		RefId:                itemInventory.ID,
		ItemInventoryContent: itemInventory.ItemInventoryContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &itemInventoryAt); err != nil {
		return errors.New("failed inserting item inventory at")
	}
	return nil
}

func UpdateItemInventory(tx *gorm.DB, basedId uint, itemInventory models.ItemInventory, at models.At, conditions map[string]interface{}) error {
	itemInventory.BasedId = basedId

	if itemInventory.ID == 0 {
		return CreateItemInventory(tx, basedId, itemInventory, at)
	}

	// Every inventory field the Item Entry form sends, blanks and zeros included (see
	// services.DbUpdateFields). A minimum lowered to 0 used to keep its old value, and REQ QTY
	// reads the minimum (spec 8.12). IsSpecialItem is left out: the form sends it as
	// special_item, not is_special_item, so it always arrives empty and would be cleared.
	if err := services.DbUpdateFields(tx, &itemInventory, conditions,
		"WarehouseId", "DefaultZone", "StorageType", "DefaultBinLocation",
		"ValuationMethodId", "MinimunInventory", "MaximunInventory",
	); err != nil {
		return errors.New("failed updating item inventory")
	}

	itemInventoryAt := models.ItemInventoryAt{
		RefId:                itemInventory.ID,
		ItemInventoryContent: itemInventory.ItemInventoryContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &itemInventoryAt); err != nil {
		return errors.New("failed inserting item inventory at")
	}
	return nil
}

func DeleteInventoryItem(tx *gorm.DB, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &models.ItemInventory{}, conditions); err != nil {
		return errors.New("failed deleting item inventory")
	}

	itemInventoryAt := models.ItemInventoryAt{
		At: models.At{},
	}

	if err := services.DbInsert(tx, &itemInventoryAt); err != nil {
		return errors.New("failed inserting item inventory at")
	}

	return nil
}
