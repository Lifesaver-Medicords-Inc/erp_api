package setup_services

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.Item
	ItemSpecs models.ItemSpecs `json:"item_specs"`
}

type SaveBody struct {
	models.Item
	ItemSpecs       []models.ItemSpecs     `json:"item_specs"`
	AdditionalSpecs models.AdditionalSpecs `json:"additional_specs"`
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items           []models.ItemView        `json:"items"`
		ItemSpecs       []models.ItemSpecs       `json:"item_specs"`
		AdditionalSpecs []models.AdditionalSpecs `json:"additional_specs"`
	}

	var response Response

	if err := services.DbGet(&response.Items, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting items")
	}

	if err := GetItemSpecs(&response.ItemSpecs, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetAdditionalSpecs(&response.AdditionalSpecs, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

// func GetItem(id int) (Body, int, error) {
// 	conditions := map[string]interface{}{
// 		"id": id,
// 	}

// 	var record Body

// 	if err := services.DbGet(&record.Item, conditions); err != nil {
// 		return record, fiber.StatusInternalServerError, errors.New("failed getting item")
// 	}

// 	conditions = map[string]interface{}{
// 		"based_id": record.Item.ID,
// 	}

// 	if err := GetItemSpec(&record.ItemSpecs, conditions); err != nil {
// 		return record, fiber.StatusInternalServerError, err
// 	}

//		return record, 0, nil
//	}
func GetItem(id int) (SaveBody, int, error) {

	conditions := map[string]interface{}{
		"id": id, // Fetch item by its unique 'id'
	}

	var record SaveBody

	// Fetch Item data
	if err := services.DbGet(&record.Item, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting item")
	}

	// Fetch ItemSpecs data
	itemSpecsConditions := map[string]interface{}{
		"based_id": record.Item.ID,
	}

	if err := GetItemSpecs(&record.ItemSpecs, itemSpecsConditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := GetAdditionalSpec(&record.AdditionalSpecs, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

// func CreateItem(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
// 	var body Body
// 	if err := c.BodyParser(&body); err != nil {
// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	if err := services.DbInsert(tx, &body.Item); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed creating item")
// 	}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	atdata := models.ItemAt{RefId: body.ID, ItemContent: body.ItemContent, At: at}

//		if err := services.DbInsert(tx, &atdata); err != nil {
//			return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
//		}
//		if err := CreateItemSpecs(tx, body.ID, body.ItemSpecs, at); err != nil {
//		// if err := CreateItemSpecs(tx, body.ID, []models.ItemSpecs{body.ItemSpecs}, at); err != nil {
//			return body, fiber.StatusInternalServerError, err
//		}
//			return body, 0, nil
//	}
func CreateItem(c *fiber.Ctx, tx *gorm.DB) (SaveBody, int, error) {
	var savebody SaveBody

	if err := c.BodyParser(&savebody); err != nil {
		return savebody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Insert the main item record into the database
	if err := services.DbInsert(tx, &savebody.Item); err != nil {
		return savebody, fiber.StatusInternalServerError, errors.New("failed creating item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Insert the associated ItemAt record
	atdata := models.ItemAt{RefId: savebody.ID, ItemContent: savebody.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return savebody, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	// Insert ItemSpecs
	if err := CreateItemSpecs(tx, savebody.ID, savebody.ItemSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	return savebody, 0, nil
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

	atdata := models.ItemAt{RefId: body.ID, ItemContent: body.ItemContent, At: at}

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
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.Item, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: body.ID, ItemContent: body.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := DeleteItemSpecs(tx, body.ItemSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
