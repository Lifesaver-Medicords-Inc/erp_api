package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPositions(conditions map[string]interface{}) ([]adminmodels.Position, int, error) {

	var based_service = services.NewInMemoryRepository(nil, nil, adminmodels.Position{}, adminmodels.PositionAt{})

	return based_service.FetchAll()
}

func GetPosition(id int) (models.Position, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var position models.Position

	if err := services.DbGet(&position, conditions); err != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed getting position")
	}

	return position, 0, nil
}
func CreatePosition(c *fiber.Ctx, tx *gorm.DB) (adminmodels.Position, int, error) {

	var service = services.NewInMemoryRepository(c, tx, adminmodels.Position{}, adminmodels.PositionAt{})

	var body adminmodels.Position
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.PositionAt{RefId: body.ID, Code: body.Name, At: at}

	return service.Create(body, atdata)
}

func UpdatePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.Position, int, error) {

	var service = services.NewInMemoryRepository(c, tx, adminmodels.Position{}, adminmodels.PositionAt{})

	var body adminmodels.Position
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := adminmodels.PositionAt{RefId: body.ID, Code: body.Name, At: at}

	return service.Update(body, atdata, conditions)
}

func DeletePosition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (adminmodels.Position, int, error) {

	var body adminmodels.Position
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

	atdata := adminmodels.PositionAt{RefId: body.ID, Code: body.Name, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
