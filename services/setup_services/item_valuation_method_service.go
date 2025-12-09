package setup_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetValuationMethods(conditions map[string]interface{}) ([]models.ValuationMethod, int, error) {
	var valuationMethods []models.ValuationMethod

	if err := services.DbGet(&valuationMethods, conditions); err != nil {
		return valuationMethods, fiber.StatusInternalServerError, errors.New("failed getting valuation methods")
	}

	return valuationMethods, 0, nil
}

func GetValuationMethod(id int) (models.ValuationMethod, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var valuationMethod models.ValuationMethod

	if err := services.DbGet(&valuationMethod, conditions); err != nil {
		return valuationMethod, fiber.StatusInternalServerError, errors.New("failed getting valuation method")
	}

	return valuationMethod, 0, nil
}

func CreateValuationMethod(c *fiber.Ctx, tx *gorm.DB) (models.ValuationMethod, int, error) {
	var body models.ValuationMethod

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating brand")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	fmt.Println("at  ok ", at)

	if !ok {
		at = models.At{}
		fmt.Println("at not ok ", at)

	}

	atdata := models.ValuationMethodAt{RefId: body.ID, Code: body.Code, ValuationMethodContent: models.ValuationMethodContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating valuationmethodat")
	}

	return body, 0, nil
}

func UpdateValuationMethod(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ValuationMethod, int, error) {
	var body models.ValuationMethod
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating brand")
	}

	at, ok := c.Locals("at").(models.At)
	fmt.Println("ATTTT", at)
	if !ok {
		fmt.Println("ATTTT11", at)

		at = models.At{}
	}

	atdata := models.ValuationMethodAt{RefId: body.ID, Code: body.Code, ValuationMethodContent: models.ValuationMethodContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating valuationmethodat")
	}

	return body, 0, nil
}

func DeleteValuationMethod(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ValuationMethod, int, error) {
	var body models.ValuationMethod
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting valuation method")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ValuationMethodAt{RefId: body.ID, Code: body.Code, ValuationMethodContent: models.ValuationMethodContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating brandat")
	}

	return body, 0, nil
}
