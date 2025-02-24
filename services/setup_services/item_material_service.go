package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetMaterials(conditions map[string]interface{}) ([]models.Material, int, error) {
	var materials []models.Material

	if err := services.DbGet(&materials, conditions); err != nil {
		return materials, fiber.StatusInternalServerError, errors.New("failed getting materials")
	}

	return materials, 0, nil
}
func GetMaterial(id int) (models.Material, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var material models.Material

	if err := services.DbGet(&material, conditions); err != nil {
		return material, fiber.StatusInternalServerError, errors.New("failed getting material")
	}

	return material, 0, nil
}

func CreateMaterial(c *fiber.Ctx, tx *gorm.DB) (models.Material, int, error) {
	var body models.Material
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating material")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.MaterialAt{RefId: body.ID, Code: body.Code, MaterialContent: body.MaterialContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating materialat")
	}

	return body, 0, nil
}
func UpdateMaterial(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Material, int, error) {
	var body models.Material
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating material")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.MaterialAt{RefId: body.ID, Code: body.Code, MaterialContent: body.MaterialContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating materialat")
	}

	return body, 0, nil
}

func DeleteMaterial(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Material, int, error) {
	var body models.Material
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting material")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.MaterialAt{RefId: body.ID, Code: body.Code, MaterialContent: body.MaterialContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating materialat")
	}

	return body, 0, nil
}
