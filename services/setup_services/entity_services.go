package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetEntities(conditions map[string]interface{}) ([]models.Entity, int, error) {
	var entity []models.Entity

	if err := services.DbGet(&entity, conditions); err != nil {
		return entity, fiber.StatusInternalServerError, errors.New("failed getting entity")
	}

	return entity, 0, nil
}

func GetEntity(id int) (models.Entity, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var entity models.Entity

	if err := services.DbGet(&entity, conditions); err != nil {
		return entity, fiber.StatusInternalServerError, errors.New("failed getting entity")
	}

	return entity, 0, nil
}

func CreateEntity(c *fiber.Ctx, tx *gorm.DB) (models.Entity, int, error) {
	var body models.Entity
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

	atdata := models.EntityAt{RefId: body.ID, Code: body.Code, EntityContent: models.EntityContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating entity at")
	}

	return body, 0, nil
}

func UpdateEntity(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Entity, int, error) {
	var body models.Entity
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating entity")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.EntityAt{RefId: body.ID, Code: body.Code, EntityContent: models.EntityContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating entity at")
	}

	return body, 0, nil
}

func DeleteEntity(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Entity, int, error) {
	var body models.Entity
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting entity")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.EntityAt{RefId: body.ID, Code: body.Code, EntityContent: models.EntityContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating social media at")
	}

	return body, 0, nil
}
