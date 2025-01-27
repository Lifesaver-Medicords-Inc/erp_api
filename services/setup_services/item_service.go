package setup_services

import (
	// "errors"

	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.Item
	ItemSpecs models.ItemSpecs `json:"item_specs"`
}

func GetItems(conditions map[string]interface{}) ([]Body, int, error) {
	var records []Body
	var items []models.Item

	if err := services.DbGet(&items, conditions); err != nil {
		return records, fiber.StatusInternalServerError, errors.New("failed getting itemspecs")
	}

	fmt.Println("Item Specsssss:", items)

	for _, v := range items {
		var itemspecs models.ItemSpecs

		conditions := map[string]interface{}{
			"based_id": v.ID,
		}

		//CHILD 1
		if err := GetItemSpecs(&itemspecs, conditions); err != nil {
			return records, fiber.StatusInternalServerError, err
		}

		body := Body{
			Item:      v,
			ItemSpecs: itemspecs,
		}

		records = append(records, body)
	}

	return records, 0, nil
}

func GetItem(id int) (Body, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var record Body

	if err := services.DbGet(&record.Item, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting item")
	}

	conditions = map[string]interface{}{
		"based_id": record.Item.ID,
	}

	//Child 1
	if err := GetItemSpecs(&record.ItemSpecs, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

func CreateItem(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.Item); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: models.ItemContent{ItemNameId: body.ItemNameId, ItemModelId: body.ItemModelId, ItemCode: body.ItemCode, ShortDesc: body.ShortDesc, LongDesc: body.LongDesc, ItemClassId: body.ItemClassId, ItemBrandId: body.ItemBrandId, UnitOfMeasureId: body.UnitOfMeasureId, IsInventoryItem: body.IsInventoryItem, IsSalesItem: body.IsSalesItem, IsPurchaseItem: body.IsPurchaseItem}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	if err := CreateItemSpecs(tx, body.ID, body.ItemSpecs, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	return body, 0, nil
}

func UpdateItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.Item, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: models.ItemContent{ItemNameId: body.ItemNameId, ItemModelId: body.ItemModelId, ItemCode: body.ItemCode, ShortDesc: body.ShortDesc, LongDesc: body.LongDesc, ItemClassId: body.ItemClassId, ItemBrandId: body.ItemBrandId, UnitOfMeasureId: body.UnitOfMeasureId, IsInventoryItem: body.IsInventoryItem, IsSalesItem: body.IsSalesItem, IsPurchaseItem: body.IsPurchaseItem}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	if err := UpdateItemSpecs(tx, body.ItemSpecs, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func DeleteItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	fmt.Println("BODYYYY:", body)
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: models.ItemContent{ItemNameId: body.ItemNameId, ItemModelId: body.ItemModelId, ItemCode: body.ItemCode, ShortDesc: body.ShortDesc, LongDesc: body.LongDesc, ItemClassId: body.ItemClassId, ItemBrandId: body.ItemBrandId, UnitOfMeasureId: body.UnitOfMeasureId, IsInventoryItem: body.IsInventoryItem, IsSalesItem: body.IsSalesItem, IsPurchaseItem: body.IsPurchaseItem}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	if err := DeleteItemSpecs(tx, body.ItemSpecs, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
