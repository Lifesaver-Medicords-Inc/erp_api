package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetUnitMeasurements() ([]models.UnitMeasurement, error) {
	var unitMeasurement []models.UnitMeasurement

	if err := services.DbGet(&unitMeasurement, nil); err != nil {
		return unitMeasurement, err
	}

	return unitMeasurement, nil

}

func CreateUnitMeasurement(c *fiber.Ctx, tx *gorm.DB) error {

	// initialize models and parse it JSON to Model
	var body models.UnitMeasurement
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	// pass the model to execute Dbinsert
	if err := services.DbInsert(tx, &body); err != nil {

		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record")
		}
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	// insert the at model
	atData := models.UnitMeasurementAt{RefId: body.ID, Code: body.Code, UnitMeasurementContent: models.UnitMeasurementContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atData); err != nil {
		return err
	}

	return nil
}

func UpdateUnitMeasurment(c *fiber.Ctx, tx *gorm.DB) error {

	var body models.UnitMeasurement
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atData := models.UnitMeasurementAt{RefId: body.ID, Code: body.Code, UnitMeasurementContent: models.UnitMeasurementContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atData); err != nil {
		return err
	}

	return nil
}

func DeleteUnitMeasurment(c *fiber.Ctx, tx *gorm.DB) error {

	var body models.UnitMeasurement
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atData := models.UnitMeasurementAt{RefId: body.ID, Code: body.Code, UnitMeasurementContent: models.UnitMeasurementContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atData); err != nil {
		return err
	}

	return nil
}
