package setup_services

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetChartGroups(conditions map[string]interface{}) ([]accounting_models.ChartGroup, int, error) {
	var based_service = services.NewInMemoryRepository(nil, nil, accounting_models.ChartGroup{}, accounting_models.ChartGroupAt{})

	return based_service.FetchAll()
}
func GetChartGroup(id int) (accounting_models.ChartGroup, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var Group accounting_models.ChartGroup

	if err := services.DbGet(&Group, conditions); err != nil {
		return Group, fiber.StatusInternalServerError, errors.New("failed getting Group")
	}

	return Group, 0, nil
}

func CreateChartGroup(c *fiber.Ctx, tx *gorm.DB) (accounting_models.ChartGroup, int, error) {
	var based_service = services.NewInMemoryRepository(c, tx, accounting_models.ChartGroup{}, accounting_models.ChartGroupAt{})

	var body accounting_models.ChartGroup
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.ChartGroupAt{RefId: body.ID, Code: body.Code, At: at}

	return based_service.Create(body, atdata)
}

func UpdateChartGroup(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (accounting_models.ChartGroup, int, error) {
	var based_service = services.NewInMemoryRepository(c, tx, accounting_models.ChartGroup{}, accounting_models.ChartGroupAt{})

	var body accounting_models.ChartGroup
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.ChartGroupAt{RefId: body.ID, Code: body.Code, At: at}

	return based_service.Update(body, atdata, conditions)
}

func DeleteChartGroup(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (accounting_models.ChartGroup, int, error) {
	var body accounting_models.ChartGroup
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting Group")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.ChartGroupAt{RefId: body.ID, Code: body.Code, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating Groupat")
	}

	return body, 0, nil
}
