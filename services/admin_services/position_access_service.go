package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPositionAccess(conditions map[string]interface{}) ([]adminmodels.PositionAccess, int, error) {

	var access []adminmodels.PositionAccess

	if err := services.DbGet(&access, conditions); err != nil {
		return access, fiber.StatusInternalServerError, errors.New("failed getting position access")
	}

	return access, 0, nil
}

func CreatePositionAccess(c *fiber.Ctx, tx *gorm.DB) (adminmodels.PositionAccess, int, error) {

	var service = services.NewInMemoryRepository(c, tx, adminmodels.PositionAccess{}, adminmodels.PositionAccessAt{})

	var body adminmodels.PositionAccess
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.PositionAccessAt{RefId: body.ID, Code: body.Code, At: at, PositionAccess: adminmodels.PositionAccess{
		PositionId: body.PositionId,
		Code:       body.Code,
	}}

	return service.Create(body, atdata)
}

func UpdatePositionAccess(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.PositionAccess, int, error) {

	var service = services.NewInMemoryRepository(c, tx, adminmodels.PositionAccess{}, adminmodels.PositionAccessAt{})

	var body adminmodels.PositionAccess
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.PositionAccessAt{RefId: body.ID, Code: body.Code, At: at, PositionAccess: adminmodels.PositionAccess{
		PositionId: body.PositionId,
		Code:       body.Code,
	}}

	return service.Update(body, atdata, conditions)
}

func DeletePositionAccess(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.PositionAccess, int, error) {

	var body adminmodels.PositionAccess
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting class")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.PositionAccessAt{RefId: body.ID, Code: body.Code, At: at, PositionAccess: adminmodels.PositionAccess{
		PositionId: body.PositionId,
		Code:       body.Code,
	}}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
