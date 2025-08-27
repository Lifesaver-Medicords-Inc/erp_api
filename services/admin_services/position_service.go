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

func GetPositions(conditions map[string]interface{}, tx *gorm.DB) ([]adminmodels.Position, int, error) {

	var positions []adminmodels.Position

	if err := services.DbGetNoCache(&positions, conditions); err != nil {
		return positions, fiber.StatusInternalServerError, errors.New("failed getting positions")
	}
	for i := range positions {
		accessCondition := map[string]interface{}{
			"position_id": positions[i].ID,
		}

		access, _, err := GetPositionAccess(accessCondition)
		if err != nil {
			continue
		}

		positions[i].Access = toPtrSlice(access)
	}

	return positions, 0, nil
}

func toPtrSlice(items []adminmodels.PositionAccess) []*adminmodels.PositionAccess {
	ptrs := make([]*adminmodels.PositionAccess, len(items))
	for i := range items {
		ptrs[i] = &items[i]
	}
	return ptrs
}

func GetPosition(conditions map[string]interface{}, tx *gorm.DB) (adminmodels.Position, int, error) {

	var position adminmodels.Position

	if err := services.DbGetNoCache(&position, conditions); err != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed getting position")
	}

	accessCondition := map[string]interface{}{
		"position_id": position.ID,
	}

	access, _, err := GetPositionAccess(accessCondition)

	if err == nil {
		accessPtrs := make([]*adminmodels.PositionAccess, len(access))
		for i := range access {
			accessPtrs[i] = &access[i]
		}
		position.Access = accessPtrs
	}

	return position, 0, nil
}

func CreatePosition(c *fiber.Ctx, tx *gorm.DB) (adminmodels.Position, int, error) {
	var body adminmodels.Position
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

	atdata := adminmodels.PositionAt{RefId: body.ID, Position: adminmodels.Position{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}

func UpdatePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.Position, int, error) {
	var body adminmodels.Position

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

	atdata := adminmodels.PositionAt{RefId: body.ID, Position: adminmodels.Position{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}

func DeletePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.Position, int, error) {

	var body adminmodels.Position
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

	atdata := adminmodels.PositionAt{RefId: body.ID, Position: adminmodels.Position{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	return body, 0, nil
}
