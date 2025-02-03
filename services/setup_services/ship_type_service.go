package setup_services

import (
	"errors"
	//fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetShipTypes(conditions map[string]interface{}) ([]models.ShipType, int, error) {
	var shiptypes []models.ShipType

	if err := services.DbGet(&shiptypes, conditions); err != nil {
		return shiptypes, fiber.StatusInternalServerError, errors.New("failed getting ship types")
	}

	return shiptypes, 0, nil
}

func GetShipType(id int) (models.ShipType, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var shiptype models.ShipType

	if err := services.DbGet(&shiptype, conditions); err != nil {
		return shiptype, fiber.StatusInternalServerError, errors.New("failed getting ship type")
	}

	return shiptype, 0, nil
}

func CreateShipType(c *fiber.Ctx, tx *gorm.DB) (models.ShipType, int, error) {
	var body models.ShipType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating shiptype")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ShipTypeAt{RefId: body.ID, ShipTypeContent: models.ShipTypeContent{ShipName: body.ShipName}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating ship type")
	}

	return body, 0, nil
}

func UpdateShipType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ShipType, int, error) {
	var body models.ShipType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating ship type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ShipTypeAt{RefId: body.ID, ShipTypeContent: models.ShipTypeContent{ShipName: body.ShipName}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating shiptypeat")
	}

	return body, 0, nil
}

func DeleteShipType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ShipType, int, error) {
	var body models.ShipType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting ship type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ShipTypeAt{RefId: body.ID, ShipTypeContent: models.ShipTypeContent{ShipName: body.ShipName}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating shiptypeat")
	}

	return body, 0, nil
}
