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

func GetClasses(conditions map[string]interface{}) ([]models.Class, int, error) {
	var classes []models.Class

	if err := services.DbGet(&classes, conditions); err != nil {
		return classes, fiber.StatusInternalServerError, errors.New("failed getting class")
	}

	return classes, 0, nil
}
func GetClass(id int) (models.Class, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var class models.Class

	if err := services.DbGet(&class, conditions); err != nil {
		return class, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return class, 0, nil
}

func CreateClass(c *fiber.Ctx, tx *gorm.DB) (models.Class, int, error) {
	var body models.Class
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating class")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ClassAt{RefId: body.ID, Code: body.Code, ClassContent: models.ClassContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
func UpdateClass(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Class, int, error) {
	var body models.Class
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating class")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ClassAt{RefId: body.ID, Code: body.Code, ClassContent: models.ClassContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}

func DeleteClass(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Class, int, error) {
	var body models.Class
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting class")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ClassAt{RefId: body.ID, Code: body.Code, ClassContent: models.ClassContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating brandat")
	}

	return body, 0, nil
}
