package position_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPositions(conditions map[string]interface{}) ([]models.PositionModel, int, error) {
	var positions []models.PositionModel

	if err := services.DbGet(&positions, conditions); err != nil {
		return positions, fiber.StatusInternalServerError, errors.New("failed getting positions")
	}

	return positions, 0, nil
}

func GetPosition(id int) (models.PositionModel, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var position models.PositionModel

	if err := services.DbGet(&position, conditions); err != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed getting position")
	}

	return position, 0, nil
}

func CreatePosition(c *fiber.Ctx, tx *gorm.DB) (models.PositionModel, int, error) {
	var body models.PositionModel
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating position")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PositionAt{RefId: body.ID, Code: body.Code, PositionContent: models.PositionContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}

func UpdatePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PositionModel, int, error) {
	var body models.PositionModel

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating position")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PositionAt{RefId: body.ID, Code: body.Code, PositionContent: models.PositionContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}

func DeletePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PositionModel, int, error) {
	var body models.PositionModel
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting position")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PositionAt{RefId: body.ID, Code: body.Code, PositionContent: models.PositionContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}
