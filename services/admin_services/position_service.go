package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPositions(conditions map[string]interface{}, tx *gorm.DB) ([]models.Position, int, error) {

	var positions []models.Position

	if err := services.DbGetNoCache(&positions, conditions); err != nil {
		return positions, fiber.StatusInternalServerError, errors.New("failed getting positions")
	}
	return positions, 0, nil
}

func ToPtrSlice(items []adminmodels.PositionAccess) []*adminmodels.PositionAccess {
	ptrs := make([]*adminmodels.PositionAccess, len(items))
	for i := range items {
		ptrs[i] = &items[i]
	}
	return ptrs
}

func GetPosition(conditions map[string]interface{}, tx *gorm.DB) (models.Position, int, error) {

	var position models.Position

	if err := services.DbGetNoCache(&position, conditions); err != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed getting position")
	}

	return position, 0, nil
}

func CreatePosition(c *fiber.Ctx, tx *gorm.DB) (models.Position, int, error) {
	var body models.Position
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

	atdata := models.PositionAt{RefId: body.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}

func UpdatePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Position, int, error) {
	var body models.Position

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

	atdata := models.PositionAt{RefId: body.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}

func DeletePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Position, int, error) {

	var body models.Position
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

	atdata := models.PositionAt{RefId: body.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}
