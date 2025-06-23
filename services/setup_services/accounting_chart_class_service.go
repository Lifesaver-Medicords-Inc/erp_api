package setup_services

import (
	// "errors"

	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetChartClasses(conditions map[string]interface{}) ([]accounting_models.ChartClass, int, error) {
	var classes []accounting_models.ChartClass

	if err := services.DbGet(&classes, conditions); err != nil {
		return classes, fiber.StatusInternalServerError, errors.New("failed getting classes")
	}

	return classes, 0, nil
}
func GetChartClass(id int) (accounting_models.ChartClass, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var class accounting_models.ChartClass

	if err := services.DbGet(&class, conditions); err != nil {
		return class, fiber.StatusInternalServerError, errors.New("failed getting class")
	}

	return class, 0, nil
}

func CreateChartClass(c *fiber.Ctx, tx *gorm.DB) (accounting_models.ChartClass, int, error) {

	var body accounting_models.ChartClass
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating class")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.ChartClassAt{RefId: body.ID, Code: body.Code, ChartClassContent: body.ChartClassContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}

func UpdateChartClass(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (accounting_models.ChartClass, int, error) {

	var body accounting_models.ChartClass
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating class")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.ChartClassAt{RefId: body.ID, Code: body.Code, ChartClassContent: body.ChartClassContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}

func DeleteChartClass(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (accounting_models.ChartClass, int, error) {

	var body accounting_models.ChartClass
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

	atdata := accounting_models.ChartClassAt{RefId: body.ID, Code: body.Code, ChartClassContent: body.ChartClassContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
