package setup_services

import (
	"errors"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
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
	ItemSpecs       models.ItemSpecs             `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecsSchema `json:"additionalspecs"`
	ItemImages      ItemImage                    `json:"itemimages"`
	ItemInventory   models.ItemInventory         `json:"iteminventory"`
}
type ItemSpecsWithTemplate struct {
	ItemSpecs      models.ItemSpecs `json:"itemspecs"`
	TemplateFields []Field          `json:"template_fields"`
}

type ItemImage struct {
	NewImages     []models.ItemImage `json:"newimages"`
	ReplaceImages []models.ItemImage `json:"replaceimages"`
	DeleteImages  []DeleteImages     `json:"deleteimages"`
}

type ReplaceImages struct {
	ImageID uint   `json:"imageid"`
	Image   string `json:"newimage"`
}

type DeleteImages struct {
	ImageID uint `json:"imageid"`
}

type ItemSpecs struct {
	Template           string       `json:"template"`
	Fields             []SpecsField `json:"fields"`
	ManufacturerOrigin string       `json:"manufacturer_origin"`
	Fla1               string       `json:"fla_1"`
	Fla2               string       `json:"fla_2"`
	Volt1              string       `json:"volt_1"`
	Volt2              string       `json:"volt_2"`
}

type SpecsField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items            []models.ItemView                        `json:"items"`
		ItemSpecs        []models.ItemSpecs                       `json:"itemspecs"`
		AdditionalSpecs  []models.AdditionalSpecsView             `json:"additionalspecs"`
		ItemImage        []models.ItemImage                       `json:"itemimages"`
		ItemPurchasing   []models.ItemPurchasingView              `json:"itempurchasing"`
		ItemSales        []models.ItemSalesView                   `json:"itemsales"`
		ItemInventory    []models.ItemInventory                   `json:"iteminventory"`
		ItemAvailableInv []models.ItemAvailableInventoryModelView `json:"itemavailableinv"`
		ItemProductions  []models.ItemProductionView              `json:"itemproduction"`
	}

	var response Response
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// Run all 9 queries in parallel
	queries := []struct {
		name string
		fn   func() error
	}{
		{"items", func() error { return services.DbGet(&response.Items, conditions) }},
		{"itemspecs", func() error { return services.DbGetRel(&response.ItemSpecs, conditions, "ItemSpecsTemplate") }},
		{"additionalspecs", func() error { return services.DbGet(&response.AdditionalSpecs, conditions) }},
		{"itemimage", func() error { return services.DbGet(&response.ItemImage, conditions) }},
		{"itempurchasing", func() error { return services.DbGet(&response.ItemPurchasing, conditions) }},
		{"itemsales", func() error { return services.DbGet(&response.ItemSales, conditions) }},
		{"iteminventory", func() error { return services.DbGet(&response.ItemInventory, conditions) }},
		{"itemavailableinv", func() error { return services.DbGet(&response.ItemAvailableInv, conditions) }},
		{"itemproductions", func() error { return services.DbGet(&response.ItemProductions, conditions) }},
	}

	errorMap := map[string]string{
		"items":            "failed getting items",
		"itemspecs":        "failed getting item spec",
		"additionalspecs":  "failed getting item additional spec",
		"itemimage":        "failed getting item image",
		"itempurchasing":   "failed getting item purchasing",
		"itemsales":        "failed getting item sales",
		"iteminventory":    "failed getting item production",
		"itemavailableinv": "failed getting item available inventory",
		"itemproductions":  "failed getting item production",
	}

	for _, q := range queries {
		wg.Add(1)
		go func(query struct {
			name string
			fn   func() error
		}) {
			defer wg.Done()
			if err := query.fn(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = errors.New(errorMap[query.name])
				}
				mu.Unlock()
			}
		}(q)
	}

	wg.Wait()

	if firstErr != nil {
		return response, fiber.StatusInternalServerError, firstErr
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
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error") //to be added: validation of duplicate fields
		} else {
			err = errors.New("failed creating item")
		}

		return savebody, fiber.StatusInternalServerError, err
	}

	at := extractAtFromContext(c)

	atdata := models.ItemAt{
		RefId:       savebody.ID,
		ItemContent: savebody.ItemContent,
		At:          at}

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
	if err := CreateItemInventory(tx, savebody.ID, savebody.ItemInventory, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()
	return savebody, 0, nil
}

func UpdateItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (SaveBody, int, error) {
	var body SaveBody
	if err := c.BodyParser(&body); err != nil {
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

	at := extractAtFromContext(c)

	atdata := models.ItemAt{
		RefId:       body.ID,
		ItemContent: body.ItemContent,
		At:          at}

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

	if err := UpdateItemInventory(tx, body.ID, body.ItemInventory, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()
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

	at := extractAtFromContext(c)

	atdata := models.ItemAt{RefId: body.ID, ItemContent: body.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := DeleteItemSpec(tx, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()

	return body, 0, nil
}

func extractAtFromContext(c *fiber.Ctx) models.At {
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	return at
}

func InvalidateItemCaches() {
	cacheKeys := []interface{}{
		models.ItemView{},
		models.AdditionalSpecsView{},
		models.ItemImage{},
		models.ItemPurchasingView{},
		models.ItemSalesView{},
		models.ItemProductionView{},
		models.InvTrackerView{},
		models.PurchaseOrderView{},
		models.InvLogbookView{},
		models.AllBomView{},
		models.AllBinLocationView{},
		models.SalesOrderViewIR{},
		models.SalesOrderViewPA{},
		models.PurchaseOrderDetailsView{},
		models.ItemAvailableInventoryModelView{},
		accounting_models.ChartOfAccountViewList{},
		accounting_models.TaxView{},
		accounting_models.TaxDetailsView{},
		accounting_models.ChartClass{},
		accounting_models.ChartClassAt{},
		accounting_models.ChartOfAccounts{},
		accounting_models.ChartOfAccountsAt{},
		accounting_models.Tax{},
		accounting_models.TaxAt{},
		accounting_models.TaxDetails{},
		accounting_models.TaxDetailsAt{},
		accounting_models.SupplierTradeView{},
		accounting_models.PaymentVoucherDetailsGet{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
