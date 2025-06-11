package setup_services

import (
	// "errors"

	"errors"
	"fmt"
	"strings"

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
	TradeTypeId     []uint                       `json:"trade_type_id"`
	ItemSpecs       ItemSpecsWrapper             `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecsSchema `json:"additionalspecs"`
	ItemImages      ItemImageUpdate              `json:"itemimages"`
}

type UpdateBody struct {
	models.Item
	TradeTypeId     []uint                       `json:"trade_type_id"`
	ItemSpecs       ItemSpecsWrapper             `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecsSchema `json:"additionalspecs"`
	ItemImages      ItemImageUpdate              `json:"itemimages"`
}
type ItemImageUpdate struct {
	NewImages     []string        `json:"newimages"`
	ReplaceImages []ReplaceImages `json:"replaceimages"`
	DeleteImages  []DeleteImages  `json:"deleteimages"`
}

type ReplaceImages struct {
	ImageID uint   `json:"imageid"`
	Image   string `json:"newimage"`
}

type DeleteImages struct {
	ImageID uint `json:"imageid"`
}

type ItemSpecsWrapper struct {
	Template           string       `json:"template"`
	Fields             []SpecsField `json:"fields"`
	ManufacturerOrigin string       `json:"manufacturer_origin"`
}

type SpecsField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items           []models.ItemView            `json:"items"`
		ItemSpecs       []models.ItemSpecs           `json:"itemspecs"`
		AdditionalSpecs []models.AdditionalSpecsView `json:"additionalspecs"`
		ItemImage       []models.ItemImage           `json:"itemimages"`
		ItemPurchasing  []models.ItemPurchasingView  `json:"itempurchasing"`
		ItemSales       []models.ItemSalesView       `json:"itemsales"`
		ItemProductions []models.ItemProductionView  `json:"itemproduction"`
	}

	var response Response

	if err := services.DbGet(&response.Items, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting items")
	}
	//child 1
	if err := GetItemSpecs(&response.ItemSpecs, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	//child 2
	if err := GetAdditionalSpecs(&response.AdditionalSpecs, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	if err := GetItemImages(&response.ItemImage, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	if err := services.DbGet(&response.ItemPurchasing, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item purchasing")
	}
	if err := services.DbGet(&response.ItemSales, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item sales")
	}
	if err := services.DbGet(&response.ItemProductions, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item production")
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
		fmt.Println("SAVING ERROR:", err)
		return savebody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &savebody.Item); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error") //to be added: validation of duplicate fields
		} else {
			err = errors.New("failed creating item")
		}

		return savebody, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ItemAt{RefId: savebody.ID, ItemContent: savebody.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return savebody, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	for _, v := range savebody.TradeTypeId {
		if err := CreateItemTradeTypes(tx, savebody.ID, uint(v), at); err != nil {
			return savebody, 0, err
		}
	}

	if err := CreateItemSpec(tx, savebody.ID, savebody.ItemSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	if err := CreateAdditionalSpec(tx, savebody.ID, savebody.AdditionalSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	if err := CreateItemImageChild(tx, savebody.ID, savebody.ItemImages.NewImages, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()
	return savebody, 0, nil
}

func UpdateItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (UpdateBody, int, error) {
	var body UpdateBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("Parsing Error:", err)

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.Item, conditions); err != nil {
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

	atdata := models.ItemAt{RefId: body.ID, ItemContent: body.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed inserting itemat")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := UpdateTradeTypes(tx, body.ID, body.TradeTypeId, at); err != nil {
		return body, 0, err
	}

	if err := UpdateItemSpec(tx, body.ID, body.ItemSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateAdditionalSpec(tx, body.AdditionalSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateItemImage(tx, body.ID, body.ItemImages, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()
	fmt.Println("UPDATE ITEM BODY: ", body)
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

	itemview := services.GetKey(models.ItemView{}, nil)
	services.InvalidateCache(itemview)

	additionspecsview := services.GetKey(models.AdditionalSpecsView{}, nil)
	services.InvalidateCache(additionspecsview)

	return body, 0, nil
}

func InvalidateItemCaches() {
	cacheKeys := []interface{}{
		models.ItemView{},
		models.AdditionalSpecsView{},
		models.ItemImage{},
		models.ItemPurchasingView{},
		models.ItemSalesView{},
		models.ItemProductionView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
