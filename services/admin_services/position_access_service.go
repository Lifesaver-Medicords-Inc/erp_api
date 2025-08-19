package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetAllPositionAccess(conditions map[string]interface{}) ([]adminmodels.PositionAccess, int, error) {

	var based_service = services.NewInMemoryRepository(nil, nil, adminmodels.PositionAccess{}, adminmodels.PositionAccessAt{})

	return based_service.FetchAll()
}

func GetPositionAccess(conditions map[string]interface{}) ([]adminmodels.PositionAccess, int, error) {

	var based_service = services.NewInMemoryRepository(nil, nil, adminmodels.PositionAccess{}, adminmodels.PositionAccessAt{})

	return based_service.FetchWithFilter(conditions)
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

	atdata := adminmodels.PositionAccessAt{RefId: body.ID, Code: body.Code, At: at}

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

	atdata := adminmodels.PositionAccessAt{RefId: body.ID, Code: body.Code, At: at}

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

	atdata := adminmodels.PositionAccessAt{RefId: body.ID, Code: body.Code, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
