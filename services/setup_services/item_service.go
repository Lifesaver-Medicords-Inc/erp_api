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

func GetItems(conditions map[string]interface{}) ([]models.Item, int, error) {
	var items []models.Item

	if err := services.DbGet(&items, conditions); err != nil {
		return items, fiber.StatusInternalServerError, errors.New("failed getting items")
	}

	return items, 0, nil
}
func GetItem(id int) (models.Item, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var item models.Item

	if err := services.DbGet(&item, conditions); err != nil {
		return item, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return item, 0, nil
}

func CreateItem(c *fiber.Ctx, tx *gorm.DB) (models.Item, int, error) {
	var body models.Item
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating item")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: models.ItemContent{ItemNameId: body.ItemNameId, ItemModelId: body.ItemModelId, ItemCode: body.ItemCode, ShortDesc: body.ShortDesc, ItemClassId: body.ShortDesc, ItemBrandId: body.ItemBrandId, UnitOfMeasureId: body.UnitOfMeasureId, IsInventoryItem: body.IsInventoryItem, IsSalesItem: body.IsSalesItem, IsPurchaseItem: body.IsPurchaseItem}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item at")
	}

	return body, 0, nil
}
func UpdateItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Item, int, error) {
	var body models.Item
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: models.ItemContent{ItemNameId: body.ItemNameId, ItemModelId: body.ItemModelId, ItemCode: body.ItemCode, ShortDesc: body.ShortDesc, ItemClassId: body.ShortDesc, ItemBrandId: body.ItemBrandId, UnitOfMeasureId: body.UnitOfMeasureId, IsInventoryItem: body.IsInventoryItem, IsSalesItem: body.IsSalesItem, IsPurchaseItem: body.IsPurchaseItem}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	return body, 0, nil
}

func DeleteItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Item, int, error) {
	var body models.Item
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting Item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: models.ItemContent{ItemNameId: body.ItemNameId, ItemModelId: body.ItemModelId, ItemCode: body.ItemCode, ShortDesc: body.ShortDesc, ItemClassId: body.ShortDesc, ItemBrandId: body.ItemBrandId, UnitOfMeasureId: body.UnitOfMeasureId, IsInventoryItem: body.IsInventoryItem, IsSalesItem: body.IsSalesItem, IsPurchaseItem: body.IsPurchaseItem}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	return body, 0, nil
}
