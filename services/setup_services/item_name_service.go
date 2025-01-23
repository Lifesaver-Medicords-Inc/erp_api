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

func GetNames() ([]models.Name, error) {
	var names []models.Name

	if err := services.DbGet(&names, nil); err != nil {
		return names, err
	}

	return names, nil
}
func GetName(id int) (models.Name, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var name models.Name

	if err := services.DbGet(&name, conditions); err != nil {
		return name, err
	}

	return name, nil
}

func CreateName(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Name
	if err := c.BodyParser(&body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record")
		}
		return err
	}

	if err := services.DbInsert(tx, &body); err != nil {
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		//return errors.New("error AT data")
		at = models.At{}
	}

	atdata := models.NameAt{RefId: body.ID, Code: body.Code, NameContent: models.NameContent{Name: body.Code}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
func UpdateName(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Name
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

	atdata := models.NameAt{RefId: body.ID, Code: body.Code, NameContent: models.NameContent{Name: body.Code}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func DeleteName(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Name
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		//return errors.New("error AT data")
		at = models.At{}
	}

	atdata := models.NameAt{RefId: body.ID, Code: body.Code, NameContent: models.NameContent{Name: body.Code}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
