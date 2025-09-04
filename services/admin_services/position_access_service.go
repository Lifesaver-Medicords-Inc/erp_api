package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPositionAccess(conditions map[string]interface{}, tx *gorm.DB) ([]models.PositionAccess, int, error) {

	var access []models.PositionAccess

	if err := tx.Where(conditions).Preload("Position").Find(&access).Error; err != nil {
		return access, fiber.StatusInternalServerError, errors.New("failed getting position access")
	}

	return access, 0, nil
}

func CreatePositionAccess(c *fiber.Ctx, tx *gorm.DB) (models.PositionAccess, int, error) {

	var body models.PositionAccess
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating position access")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PositionAccessAt{RefId: body.ID, Code: body.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: body.PositionId,
		Code:       body.Code,
	}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionaccessat")
	}

	return body, 0, nil
}

func UpdatePositionAccess(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PositionAccess, int, error) {
	var body models.PositionAccess

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating position access")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PositionAccessAt{RefId: body.ID, Code: body.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: body.PositionId,
		Code:       body.Code,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionaccessat")
	}

	return body, 0, nil
}

func DeletePositionAccess(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PositionAccess, int, error) {

	var body models.PositionAccess
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting position access")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PositionAccessAt{RefId: body.ID, Code: body.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: body.PositionId,
		Code:       body.Code,
	}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionaccessat")
	}

	return body, 0, nil
}
