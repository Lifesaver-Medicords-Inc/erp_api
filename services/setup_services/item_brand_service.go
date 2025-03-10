package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetBrands(conditions map[string]interface{}) ([]models.Brand, int, error) {
	var brands []models.Brand

	if err := services.DbGet(&brands, conditions); err != nil {
		return brands, fiber.StatusInternalServerError, errors.New("failed getting brands")
	}

	return brands, 0, nil
}

func GetBrand(id int) (models.Brand, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var brand models.Brand

	if err := services.DbGet(&brand, conditions); err != nil {
		return brand, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return brand, 0, nil
}

func CreateBrand(tx *gorm.DB, brand *models.Brand) (int, error) {
	if err := services.DbInsert(tx, brand); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating brand")
		}

		return fiber.StatusInternalServerError, err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	at = models.At{}
	// }

	// atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: at}
	// if err := services.DbInsert(tx, &atdata); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed creating brandat")
	// }

	return 0, nil
}

func UpdateBrand(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Brand, int, error) {
	var body models.Brand
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating brand")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating brandat")
	}

	return body, 0, nil
}

func DeleteBrand(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Brand, int, error) {
	var body models.Brand
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting brand")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.BrandAt{RefId: body.ID, Code: body.Code, BrandContent: models.BrandContent{Name: body.Name}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating brandat")
	}

	return body, 0, nil
}
