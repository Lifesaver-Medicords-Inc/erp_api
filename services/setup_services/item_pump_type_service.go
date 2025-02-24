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

func GetPumpTypes(conditions map[string]interface{}) ([]models.PumpType, int, error) {
	var pumptype []models.PumpType

	if err := services.DbGet(&pumptype, conditions); err != nil {
		return pumptype, fiber.StatusInternalServerError, errors.New("failed getting pumptypes")
	}

	return pumptype, 0, nil
}
func GetPumpType(id int) (models.PumpType, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var pumptype models.PumpType

	if err := services.DbGet(&pumptype, conditions); err != nil {
		return pumptype, fiber.StatusInternalServerError, errors.New("failed getting pumptype")
	}

	return pumptype, 0, nil
}

func CreatePumpType(c *fiber.Ctx, tx *gorm.DB) (models.PumpType, int, error) {
	var body models.PumpType
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating pumptype")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PumpTypeAt{RefId: body.ID, Code: body.Code, PumpTypeContent: body.PumpTypeContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		fmt.Println("Error:", err)
		return body, fiber.StatusInternalServerError, errors.New("failed creating pumptypeat")
	}

	return body, 0, nil
}
func UpdatePumpType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PumpType, int, error) {
	var body models.PumpType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pumptype")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PumpTypeAt{RefId: body.ID, Code: body.Code, PumpTypeContent: body.PumpTypeContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pump typeat")
	}

	return body, 0, nil
}

func DeletePumpType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PumpType, int, error) {
	var body models.PumpType
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting pumptype")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PumpTypeAt{RefId: body.ID, Code: body.Code, PumpTypeContent: models.PumpTypeContent{}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pumptypeat")
	}

	return body, 0, nil
}
