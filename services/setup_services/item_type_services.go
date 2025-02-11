package setup_services

import (
	// "errors"

	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetTypes(conditions map[string]interface{}) ([]models.Type, int, error) {
	var types []models.Type

	if err := services.DbGet(&types, conditions); err != nil {
		return types, fiber.StatusInternalServerError, errors.New("failed getting types")
	}

	return types, 0, nil
}
func GetType(id int) (models.Type, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var itemType models.Type

	if err := services.DbGet(&itemType, conditions); err != nil {
		return itemType, fiber.StatusInternalServerError, errors.New("failed getting type")
	}

	return itemType, 0, nil
}

func CreateType(c *fiber.Ctx, tx *gorm.DB) (models.Type, int, error) {
	var body models.Type
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating type")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.TypeAt{RefId: body.ID, Code: body.Code, TypeContent: body.TypeContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating typeat")
	}

	return body, 0, nil
}

func UpdateType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Type, int, error) {
	var body models.Type
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.TypeAt{RefId: body.ID, Code: body.Code, TypeContent: body.TypeContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating typeat")
	}

	return body, 0, nil
}

func DeleteType(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Type, int, error) {
	var body models.Type
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting type")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.TypeAt{RefId: body.ID, Code: body.Code, TypeContent: body.TypeContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating typeat")
	}

	return body, 0, nil
}

