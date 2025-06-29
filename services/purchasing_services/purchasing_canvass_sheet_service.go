package purchasing_services

import (
	"errors"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/bpi_services"
	"gorm.io/gorm"
)

type Body struct {
	models.PurchasingCanvassSheet
}

func GetPurchasingCanvasSheetSO(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.PurchasingCanvassSheetSOView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasingcanvasssheet")
	}

	// sort by current_list_price after query
	sort.Slice(response, func(i, j int) bool {
		return response[i].CurrentListPrice < response[j].CurrentListPrice
	})

	return response, 0, nil
}

func CreatePurchasingCanvassSheet(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.PurchasingCanvassSheet); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasingcanvasssheet")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	parentat := models.PurchasingCanvassSheetAt{
		RefId:                         body.ID,
		PurchasingCanvassSheetContent: body.PurchasingCanvassSheetContent,
		At:                            at}

	if err := services.DbInsert(tx, &parentat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasingcanvasssheetat")
	}

	conditions := map[string]interface{}{
		"item_id":   body.ItemId,
		"branch_id": body.SupplierId,
	}

	if err := bpi_services.UpdateBpiItemCanvass(tx, body.ID, body.NetPrice, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	InvalidateItemCaches()

	return body, 0, nil
}

func UpdatePurchasingCanvassSheet(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.PurchasingCanvassSheet, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating purchasingcanvasssheet")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PurchasingCanvassSheetAt{
		RefId:                         body.ID,
		PurchasingCanvassSheetContent: body.PurchasingCanvassSheetContent,
		At:                            at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasingcanvasssheetat")
	}

	conditions = map[string]interface{}{
		"item_id":   body.ItemId,
		"branch_id": body.SupplierId,
	}

	if err := bpi_services.UpdateBpiItemCanvass(tx, body.ID, body.NetPrice, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()

	return body, 0, nil

}

func DeletePurchasingCanvassSheetSupplier(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting supplier")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PurchasingCanvassSheetAt{RefId: body.ID, PurchasingCanvassSheetContent: body.PurchasingCanvassSheetContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasingcanvasssheetat")
	}

	InvalidateItemCaches()

	return body, 0, nil
}

func InvalidateItemCaches() {
	cacheKeys := []interface{}{
		models.PurchasingCanvassSheetSOView{},
		[]models.BpiItemsView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
