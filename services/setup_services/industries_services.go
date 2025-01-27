package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetIndustries(conditions map[string]interface{}) ([]models.Industries, int, error) {
	var industries []models.Industries

	if err := services.DbGet(&industries, conditions); err != nil {
		return industries, fiber.StatusInternalServerError, errors.New("failed getting industries")
	}

	return industries, 0, nil
}

func GetIndustry(id int) (models.Industries, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var industry models.Industries

	if err := services.DbGet(&industry, conditions); err != nil {
		return industry, fiber.StatusInternalServerError, errors.New("failed getting industries")
	}

	return industry, 0, nil
}

func CreateIndustry(c *fiber.Ctx, tx *gorm.DB) (models.Industries, int, error) {
	var body models.Industries
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating entity")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.IndustriesAt{RefId: body.ID, Code: body.Code, IndustriesContent: models.IndustriesContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating industries at")
	}

	return body, 0, nil
}

func UpdateIndustry(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Entity, int, error) {
	var body models.Entity
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating industries")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.IndustriesAt{RefId: body.ID, Code: body.Code, IndustriesContent: models.IndustriesContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating industries at")
	}

	return body, 0, nil
}

func DeleteIndustry(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Entity, int, error) {
	var body models.Entity
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting industries")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.IndustriesAt{RefId: body.ID, Code: body.Code, IndustriesContent: models.IndustriesContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating industries at")
	}

	return body, 0, nil
}
