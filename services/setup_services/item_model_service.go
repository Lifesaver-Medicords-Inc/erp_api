package setup_services

import (
	// "errors"

	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// func GetModels(conditions map[string]interface{}) ([]models.Model, int, error) {
// 	var models []models.Model

// 	if err := services.DbGet(&models, conditions); err != nil {
// 		return models, fiber.StatusInternalServerError, errors.New("failed getting models")
// 	}

//		return models, 0, nil
//	}

func GetModels(conditions map[string]interface{}) ([]models.Model, int, error) {
	// Struct to hold the joined result

	var results []models.Model

	// Access the database instance directly from initializers
	query := initializers.DB.Table("tbl_setup_item_model a").
		Select("a.*, b.name AS related_name,  c.name AS related_brand").
		Joins("LEFT JOIN tbl_setup_item_name b ON a.item_name_id = b.id").
		Joins("LEFT JOIN tbl_setup_item_brand c ON a.item_brand_id = c.id")

	// Apply conditions dynamically
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	// Execute the query
	if err := query.Find(&results).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed to get models")
	}

	// Map the results back to models.Model if necessary

	return results, 0, nil
}

func GetModel(id int) (models.Model, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var model models.Model

	if err := services.DbGet(&model, conditions); err != nil {
		return model, fiber.StatusInternalServerError, errors.New("failed getting model")
	}

	return model, 0, nil
}

func CreateModel(c *fiber.Ctx, tx *gorm.DB) (models.Model, int, error) {
	var body models.Model
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating model")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ModelAt{RefId: body.ID, ModelContent: models.ModelContent{Name: body.Name, ItemNameId: body.ItemNameId, ItemBrandId: body.ItemNameId, IsActive: body.IsActive}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating modelat")
	}

	return body, 0, nil
}
func UpdateModel(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Model, int, error) {
	var body models.Model
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating model")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ModelAt{RefId: body.ID, ModelContent: models.ModelContent{Name: body.Name, ItemNameId: body.ItemNameId, ItemBrandId: body.ItemNameId, IsActive: body.IsActive}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating modelat")
	}

	return body, 0, nil
}

func DeleteModel(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Model, int, error) {
	var body models.Model
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting model")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ModelAt{RefId: body.ID, ModelContent: models.ModelContent{Name: body.Name, ItemNameId: body.ItemNameId, ItemBrandId: body.ItemNameId, IsActive: body.IsActive}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating modelat")
	}

	return body, 0, nil
}
