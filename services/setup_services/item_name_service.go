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

func GetNames(conditions map[string]interface{}) ([]models.Name, int, error){
	var names []models.Name

	if err := services.DbGet(&names, conditions); err != nil {
		return names, fiber.StatusInternalServerError, errors.New("failed getting names")
	}

	return names, 0, nil
}

func GetName(id int) (models.Name, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var name models.Name

	if err := services.DbGet(&name, conditions); err != nil {
		return name, fiber.StatusInternalServerError, errors.New("failed getting name")
	}

	return name, 0, nil
}

func CreateName(c *fiber.Ctx, tx *gorm.DB) (models.Name, int, error) {
	var body models.Name
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating name")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.NameAt{RefId: body.ID, Code: body.Code, NameContent: models.NameContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating nameat")
	}

	return body, 0, nil
}

func UpdateName(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Name, int, error) {
	var body models.Name
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating name")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.NameAt{RefId: body.ID, Code: body.Code, NameContent: models.NameContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating nameat")
	}

	return body, 0, nil
}

func DeleteName(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Name, int, error) {
	var body models.Name
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting name")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.NameAt{RefId: body.ID, Code: body.Code, NameContent: models.NameContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating nameat")
	}

	return body, 0, nil
}

