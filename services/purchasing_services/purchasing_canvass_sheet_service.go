package purchasing_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.PurchasingCanvassSheet
}

func GetPurchasingCanvasSheet(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.PurchasingCanvassSheet

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasing canvass sheet")
	}
	return response, 0, nil
}

func CreatePurchasingCanvassSheet(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.PurchasingCanvassSheet); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasing canvass sheet")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	parentat := models.PurchasingCanvassSheetAt{RefId: body.ID, PurchasingCanvassSheetContent: body.PurchasingCanvassSheetContent, At: at}
	if err := services.DbInsert(tx, &parentat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasing canvass sheet at")
	}
	return body, 0, nil
}

func UpdatePurchasingCanvassSheet(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("UPDATE BODY:", body)
		fmt.Println("UPDATE ERR:", err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.PurchasingCanvassSheet, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating purchasing canvass sheet")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PurchasingCanvassSheetAt{RefId: body.ID, PurchasingCanvassSheetContent: body.PurchasingCanvassSheetContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchasing canvass sheet at")
	}

	return body, 0, nil
}
