package setup_services

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetModels(conditions map[string]interface{}) ([]models.Model, int, error) {
	var models []models.Model
	if err := services.DbGet(&models, conditions); err != nil {
		return models, fiber.StatusInternalServerError, errors.New("failed getting brands")
	}
	return models, 0, nil
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

func CreateModel(tx *gorm.DB, basedId uint, itemModel models.Model, at models.At) error {
	itemModel.BasedId = basedId

	if err := services.DbInsert(tx, &itemModel); err != nil {
		return errors.New("failed creating item model")
	}

	itemmodelat := models.ModelAt{
		RefId:        itemModel.ID,
		ModelContent: itemModel.ModelContent,
		At:           at,
	}

	if err := services.DbInsert(tx, &itemmodelat); err != nil {
		return errors.New("failed creating item model at")
	}

	return nil
}

// func UpdateModel(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Model, int, error) {
// 	var body models.Model
// 	if err := c.BodyParser(&body); err != nil {
// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	if err := services.DbUpdate(tx, &body, conditions); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed updating model")
// 	}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	atdata := models.ModelAt{RefId: body.ID, ModelContent: body.ModelContent, At: at}
// 	if err := services.DbInsert(tx, &atdata); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed creating modelat")
// 	}

//		return body, 0, nil
//	}

func UpdateModel(tx *gorm.DB, itemmodel models.Model, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &itemmodel, conditions); err != nil {
		return errors.New("failed updating item model")
	}

	itemmodelat := models.ModelAt{
		RefId:        itemmodel.ID,
		ModelContent: itemmodel.ModelContent,
		At:           at,
	}
	if err := services.DbInsert(tx, &itemmodelat); err != nil {
		return errors.New("failed creating item model at")
	}

	return nil
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

	atdata := models.ModelAt{RefId: body.ID, ModelContent: body.ModelContent, At: at}

	key := services.GetKey(models.ModelView{}, nil)
	services.InvalidateCache(key)

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating modelat")
	}

	return body, 0, nil
}
