package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetWarehouseAreas(WarehouseArea *models.WarehouseArea, conditions map[string]interface{}) error {
	if err := services.DbGet(WarehouseArea, conditions); err != nil {
		return errors.New("failed getting warehouse areas")
	}

	return nil
}

func GetWarehouseArea(WarehouseArea *models.WarehouseArea, conditions map[string]interface{}) error {
	if err := services.DbGet(WarehouseArea, conditions); err != nil {
		return errors.New("failed getting warehouse area")
	}

	return nil
}

func CreateWarehouseArea(tx *gorm.DB, parentId uint, child models.WarehouseArea, at models.At) error {
	content := models.WarehouseAreaContent{
		WarehouseNameId: parentId,
		UseType:         child.UseType,
		Zone:            child.Zone,
		Area:            child.Area,
		Rack:            child.Rack,
		Level:           child.Level,
		Bins:            child.Bins,
		LocationCode:    child.LocationCode,
		Notes:           child.Notes,
	}
	WarehouseArea := models.WarehouseArea{
		WarehouseAreaContent: content,
	}
	if err := services.DbInsert(tx, &WarehouseArea); err != nil {
		return errors.New("failed creating warehouse area")
	}

	WarehouseAreaAt := models.WarehouseAreaAt{
		RefId:                WarehouseArea.ID,
		WarehouseAreaContent: content,
		At:                   at,
	}
	if err := services.DbInsert(tx, &WarehouseAreaAt); err != nil {
		return errors.New("failed creating warehouse area at")
	}

	return nil
}

func UpdateWarehouseArea(tx *gorm.DB, WarehouseArea models.WarehouseArea, at models.At, conditions map[string]interface{}) error {

	if err := services.DbUpdate(tx, &WarehouseArea, conditions); err != nil {
		return errors.New("failed updating warehouse area")
	}

	WarehouseAreaAt := models.WarehouseAreaAt{
		RefId:                WarehouseArea.ID,
		WarehouseAreaContent: WarehouseArea.WarehouseAreaContent,
		At:                   at,
	}

	if err := services.DbInsert(tx, &WarehouseAreaAt); err != nil {
		return errors.New("failed creating warehouse area at")
	}

	return nil
}

func DeleteWarehouseArea(tx *gorm.DB, WarehouseArea models.WarehouseArea, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &WarehouseArea, conditions); err != nil {
		return errors.New("failed deleting warehouse area")
	}

	WarehouseAreaAt := models.WarehouseAreaAt{
		RefId:                WarehouseArea.ID,
		WarehouseAreaContent: WarehouseArea.WarehouseAreaContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &WarehouseAreaAt); err != nil {
		return errors.New("failed creating warehouse area at")
	}

	return nil
}

/// WORKING / THE USED METHODS ///
//pls dont code review :}  //note(sumabog -> nag bandaid -> pinabayaan)
//per row crud

func GetWarehouseAreasRow(conditions map[string]interface{}) ([]models.WarehouseArea, int, error) {
	var warehouseAreas []models.WarehouseArea

	if err := services.DbGet(&warehouseAreas, conditions); err != nil {
		return warehouseAreas, fiber.StatusInternalServerError, errors.New("failed getting Warehouse Areas")
	}

	return warehouseAreas, 0, nil
}

func GetWarehouseAreaRow(id int) (models.WarehouseArea, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var warehouseArea models.WarehouseArea

	if err := services.DbGet(&warehouseArea, conditions); err != nil {
		return warehouseArea, fiber.StatusInternalServerError, errors.New("failed getting warehouse areas")
	}

	return warehouseArea, 0, nil
}

func CreateWarehouseAreaRow(tx *gorm.DB, warehouseArea *models.WarehouseArea) (int, error) {

	if err := services.DbInsert(tx, warehouseArea); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating warehouse area")
		}

		return fiber.StatusInternalServerError, err
	}

	return 0, nil
}

func UpdateWarehouseAreaRow(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WarehouseArea, int, error) {
	var body models.WarehouseArea
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating warehouse area")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	WarehouseAreaAt := models.WarehouseAreaAt{
		RefId:                body.ID,
		WarehouseAreaContent: body.WarehouseAreaContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &WarehouseAreaAt); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating warehouse area")
	}

	return body, 0, nil
}

func DeleteWarehouseAreaRow(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WarehouseArea, int, error) {
	var body models.WarehouseArea
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting warehouse area")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	WarehouseAreaAt := models.WarehouseAreaAt{
		RefId:                body.ID,
		WarehouseAreaContent: body.WarehouseAreaContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &WarehouseAreaAt); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed at creating warehouse area at")
	}

	return body, 0, nil
}
