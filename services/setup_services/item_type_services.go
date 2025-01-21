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

func GetTypes() ([]models.Type, error) {
	var types []models.Type

	if err := services.DbGet(&types, nil); err != nil {
		return types, err
	}

	return types, nil
}
func GetType(id int) (models.Type, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var itemType models.Type

	if err := services.DbGet(&itemType, conditions); err != nil {
		return itemType, err
	}

	return itemType, nil
}

func CreateType(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Type
	if err := c.BodyParser(&body); err != nil {

		return err
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record")
		}
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		//return errors.New("error AT data")
		at = models.At{}
	}

	atdata := models.TypeAt{RefId: body.ID, Code: body.Code, TypeContent: models.TypeContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
func UpdateType(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Type
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		//return errors.New("error AT data")
		at = models.At{}
	}

	atdata := models.TypeAt{RefId: body.ID, Code: body.Code, TypeContent: models.TypeContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func DeleteType(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Type
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.TypeAt{RefId: body.ID, Code: body.Code, TypeContent: models.TypeContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
