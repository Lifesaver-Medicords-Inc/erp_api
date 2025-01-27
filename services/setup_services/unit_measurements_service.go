package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetUnitMeasurements(conditions map[string]interface{}) ([]models.UnitMeasurement, int, error) {
	var unitMeasurement []models.UnitMeasurement

	if err := services.DbGet(&unitMeasurement, conditions); err != nil {
		return unitMeasurement, fiber.StatusInternalServerError, errors.New("failed getting unit of measurement")
	}

	return unitMeasurement, 0, nil
}

func GetUnitMeasurement(id int) (models.UnitMeasurement, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var unitMeasurement models.UnitMeasurement

	if err := services.DbGet(&unitMeasurement, conditions); err != nil {
		return unitMeasurement, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return unitMeasurement, 0, nil
}

func CreateUnitMeasurement(c *fiber.Ctx, tx *gorm.DB) (models.UnitMeasurement, int, error) {
	var body models.UnitMeasurement
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating unit of measurement")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.UnitMeasurementAt{RefId: body.ID, Code: body.Code, UnitMeasurementContent: models.UnitMeasurementContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating unit of measurement at")
	}

	return body, 0, nil
}

func UpdateUnitMeasurement(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.UnitMeasurement, int, error) {
	var body models.UnitMeasurement
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating unit of measurement")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.UnitMeasurementAt{RefId: body.ID, Code: body.Code, UnitMeasurementContent: models.UnitMeasurementContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating unit of measurement")
	}

	return body, 0, nil
}

func DeleteUnitMeasurement(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.UnitMeasurement, int, error) {
	var body models.UnitMeasurement
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting unit of measurement")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.UnitMeasurementAt{RefId: body.ID, Code: body.Code, UnitMeasurementContent: models.UnitMeasurementContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating unt of measurement at")
	}

	return body, 0, nil
}
