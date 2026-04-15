package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetUseTypes(conditions map[string]interface{}) ([]models.WarehouseUseType, int, error) {
	var usetypes []models.WarehouseUseType

	if err := services.DbGet(&usetypes, conditions); err != nil {
		return usetypes, fiber.StatusInternalServerError, errors.New("failed getting use types")
	}

	return usetypes, 0, nil
}

func GetUseType(id int) (models.WarehouseUseType, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var usetype models.WarehouseUseType

	if err := services.DbGet(&usetype, conditions); err != nil {
		return usetype, fiber.StatusInternalServerError, errors.New("failed getting use types")
	}

	return usetype, 0, nil
}

func CreateUseType(tx *gorm.DB, usetype *models.WarehouseUseType) (int, error) {
	if err := services.DbInsert(tx, usetype); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating use type")
		}

		return fiber.StatusInternalServerError, err
	}

	return 0, nil
}

func UpdateUseType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WarehouseUseType, int, error) {
	var body models.WarehouseUseType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating use type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WarehouseUseTypeAt{RefID: body.ID, Code: body.Code, WarehouseUseTypeContent: models.WarehouseUseTypeContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating use type")
	}

	return body, 0, nil
}

func DeleteUseType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.WarehouseUseType, int, error) {
	var body models.WarehouseUseType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting use type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.WarehouseUseTypeAt{RefID: body.ID, Code: body.Code, WarehouseUseTypeContent: models.WarehouseUseTypeContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed at creating usetypeat")
	}

	return body, 0, nil
}
