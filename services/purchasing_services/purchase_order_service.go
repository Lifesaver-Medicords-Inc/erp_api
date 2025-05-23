package purchasing_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type PurchaseOrderBody struct {
	models.PurchaseOrder
	PurchaseOrderDetails []models.PurchaseOrderDetails `json:"purchase_order_details"`
}

func GetPurchaseOrder(conditions map[string]interface{}) (interface{}, int, error) {

	type Response struct {
		PurchaseOrder        []models.PurchaseOrder        `json:"purchaseorder"`
		PurchaseOrderDetails []models.PurchaseOrderDetails `json:"purchaseorderdetails"`
	}

	var response Response

	if err := services.DbGet(&response.PurchaseOrder, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchase order")
	}

	if err := GetPurchaseOrderDetails(&response.PurchaseOrderDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func CreatePurchaseOrder(c *fiber.Ctx, tx *gorm.DB) (PurchaseOrderBody, int, error) {
	var body PurchaseOrderBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("ERROR", err)
		fmt.Println("PO INSERT BODY", err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.PurchaseOrder); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchaseorder")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	purchaseorderat := models.PurchaseOrderAt{
		RefId:                body.ID,
		PurchaseOrderContent: body.PurchaseOrderContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &purchaseorderat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchaseorderat")
	}

	for _, v := range body.PurchaseOrderDetails {
		if err := CreatePurchaseOrderDetails(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}
