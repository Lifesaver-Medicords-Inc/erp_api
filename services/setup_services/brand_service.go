package setup_services

import (
	"errors"
	"strings"

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
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record")
		}

		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record")
		}
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func UpdateBrand(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Brand
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func DeleteBrand(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Brand
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: models.At{}}

	// atdata := models.BrandAt{
	// 	RefId: body.ID,
	// 	Code:  body.Code,
	// 	Name:  body.Name,
	// 	At:    at,
	// }

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
