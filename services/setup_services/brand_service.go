package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetBrands() ([]models.Brand, error) {
	var brands []models.Brand

	if err := services.DbGet(&brands, nil); err != nil {
		return brands, err
	}

	return brands, nil
}

func GetBrand(id int) (models.Brand, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var brand models.Brand

	if err := services.DbGet(&brand, conditions); err != nil {
		return brand, err
	}

	return brand, nil
}

func CreateBrand(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Brand
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbInsert(tx, &body); err != nil {
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		return errors.New("error AT data")
	}

	atdata := models.BrandAt{
		Brand: body,
		At:    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

// func Update(a int, b int) int {
// 	return a * b
// }

// func Delete(a int, b int) int {
// 	return a * b
// }
