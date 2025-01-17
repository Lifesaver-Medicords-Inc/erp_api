package setup_services

import (
	// "errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetClasses() ([]models.Class, error) {
	var class []models.Class

	if err := services.DbGet(&class, nil); err != nil {
		return class, err
	}

	return class, nil
}
func GetClass(id int) (models.Class, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var class models.Class

	if err := services.DbGet(&class, conditions); err != nil {
		return class, err
	}

	return class, nil
}

func CreateClass(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Class
	if err := c.BodyParser(&body); err != nil {
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

	atdata := models.ClassAt{
		RefId: body.ID,
		Code:  body.Code,
		Name:  body.Name,
		At:    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
func UpdateClass(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Class
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

	atdata := models.ClassAt{
		RefId: body.ID,
		Code:  body.Code,
		Name:  body.Name,
		At:    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func DeleteClass(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Class
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

	atdata := models.ClassAt{
		RefId: body.ID,
		Code:  body.Code,
		Name:  body.Name,
		At:    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
