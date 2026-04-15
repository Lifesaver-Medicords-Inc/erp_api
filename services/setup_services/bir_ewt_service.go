package setup_services

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetExpandedWithholdingTax(conditions map[string]interface{}) ([]models.ExpandedWithholdingTax, int, error) {
	var based_service = services.NewInMemoryRepository(nil, nil, models.ExpandedWithholdingTax{}, models.ExpandedWithholdingTaxAt{})

	return based_service.FetchAll()
}

func CreateExpandedWithholdingTax(c *fiber.Ctx, tx *gorm.DB) (models.ExpandedWithholdingTax, int, error) {
	var based_service = services.NewInMemoryRepository(c, tx, models.ExpandedWithholdingTax{}, models.ExpandedWithholdingTaxAt{})

	var body models.ExpandedWithholdingTax
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ExpandedWithholdingTaxAt{RefId: body.ID, Code: body.Code, At: at}

	return based_service.Create(body, atdata)
}

func UpdateExpandedWithholdingTax(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ExpandedWithholdingTax, int, error) {
	var based_service = services.NewInMemoryRepository(c, tx, models.ExpandedWithholdingTax{}, models.ExpandedWithholdingTaxAt{})

	var body models.ExpandedWithholdingTax
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ExpandedWithholdingTaxAt{RefId: body.ID, Code: body.Code, At: at}

	return based_service.Update(body, atdata, conditions)
}

func DeleteExpandedWithholdingTax(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ExpandedWithholdingTax, int, error) {
	var body models.ExpandedWithholdingTax
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting expanded tax")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ExpandedWithholdingTaxAt{RefId: body.ID, Code: body.Code, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating expanded tax")
	}

	return body, 0, nil
}
