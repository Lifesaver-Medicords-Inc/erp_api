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
	ItemSpecs       models.ItemSpecs       `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecs `json:"additional_specs"`
}

type SaveBody struct {
	models.Item
	TradeType       []string               `json:"trade_status"`
	ItemSpecs       ItemSpecsWrapper       `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecs `json:"additionalspecs"`
	// ItemPurchasing  models.ItemPurchasing  `json:"itempurchasing"`
}

type ItemSpecsWrapper struct {
	Template string       `json:"template"`
	Fields   []SpecsField `json:"fields"`
}

type SpecsField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items           []models.ItemView           `json:"items"`
		ItemSpecs       []models.ItemSpecs          `json:"itemspecs"`
		AdditionalSpecs []models.AdditionalSpecs    `json:"additionalspecs"`
		ItemPurchasing  []models.ItemPurchasingView `json:"itempurchasing"`
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
	if err := services.DbGet(&response.ItemPurchasing, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item purchasing")
	}

	return response, 0, nil
}

func GetItem(id int) (Body, int, error) {

	conditions := map[string]interface{}{
		"id": id,
	}
	var record Body

	if err := services.DbGet(&record.Item, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting item")
	}

	itemSpecsConditions := map[string]interface{}{
		"based_id": record.Item.ID,
	}

	if err := GetItemSpec(&record.ItemSpecs, itemSpecsConditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := GetAdditionalSpec(&record.AdditionalSpecs, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

func CreateItem(c *fiber.Ctx, tx *gorm.DB) (SaveBody, int, error) {
	var savebody SaveBody

	if err := c.BodyParser(&savebody); err != nil {
		return savebody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &savebody.Item); err != nil {
		return savebody, fiber.StatusInternalServerError, errors.New("failed creating item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: savebody.ID, ItemContent: savebody.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return savebody, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	key := services.GetKey(models.ItemView{}, nil)
	services.InvalidateCache(key)

	for _, v := range savebody.TradeType {
		if err := CreateTradeType(tx, savebody.ID, string(v), at); err != nil {
			return savebody, fiber.StatusInternalServerError, err
		}
	}

	if err := CreateItemSpec(tx, savebody.Item.ID, savebody.ItemSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	if err := CreateAdditionalSpec(tx, savebody.ID, savebody.AdditionalSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	// if err := CreateItemPurchasing(tx, savebody.ID, savebody.ItemPurchasing, at); err != nil {
	// 	return savebody, fiber.StatusInternalServerError, err
	// }

	return savebody, 0, nil
}

func UpdateItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (SaveBody, int, error) {
	var body SaveBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("Parsing Error:", err)

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
		return body, fiber.StatusInternalServerError, errors.New("failed inserting itemat")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := UpdateTradeTypes(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateItemSpec(tx, body.ID, body.ItemSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateAdditionalSpec(tx, body.AdditionalSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	key := services.GetKey(models.ItemView{}, nil)
	services.InvalidateCache(key)

	return body, 0, nil
}

func UpdateTradeTypes(tx *gorm.DB, body SaveBody, at models.At) error {
	if err := services.DbDelete(tx, &models.TradeType{}, map[string]interface{}{"based_id": body.ID}); err != nil {
		return errors.New("failed deleting existing trade types")
	}

	for _, v := range body.TradeType {
		if err := CreateTradeType(tx, body.ID, v, at); err != nil {
			return err
		}
	}

	return nil
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
	//additionalspecs, trade type

	return body, 0, nil
}
