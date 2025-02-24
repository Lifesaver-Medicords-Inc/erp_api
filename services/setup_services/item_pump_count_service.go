package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPumpCounts(conditions map[string]interface{}) ([]models.PumpCount, int, error) {
	var pumpcount []models.PumpCount

	if err := services.DbGet(&pumpcount, conditions); err != nil {
		return pumpcount, fiber.StatusInternalServerError, errors.New("failed getting pumpcount")
	}

	return pumpcount, 0, nil
}
func GetPumpCount(id int) (models.PumpCount, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var pumpcount models.PumpCount

	if err := services.DbGet(&pumpcount, conditions); err != nil {
		return pumpcount, fiber.StatusInternalServerError, errors.New("failed getting class")
	}

	return pumpcount, 0, nil
}

func CreatePumpCount(c *fiber.Ctx, tx *gorm.DB) (models.PumpCount, int, error) {
	var body models.PumpCount
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating pump count")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PumpCountAt{RefId: body.ID, Code: body.Code, PumpCountContent: body.PumpCountContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pumpcountat")
	}

	return body, 0, nil
}
func UpdatePumpCount(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PumpCount, int, error) {
	var body models.PumpCount
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pumpcount")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PumpCountAt{RefId: body.ID, Code: body.Code, PumpCountContent: body.PumpCountContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pumpcountat")
	}

	return body, 0, nil
}

func DeletePumpCount(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PumpCount, int, error) {
	var body models.PumpCount
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting pump type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PumpCountAt{RefId: body.ID, Code: body.Code, PumpCountContent: body.PumpCountContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pumpcountat")
	}

	return body, 0, nil
}
