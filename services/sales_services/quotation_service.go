package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.SalesQuotation
	QuickQuote models.SalesQuotationQuick `json:"sales_quotation_quick"`
}

func GetSalesQuotations(conditions map[string]interface{}) ([]Body, int, error) {
	var records []Body
	var quotations []models.SalesQuotation

	if err := services.DbGet(&quotations, conditions); err != nil {
		return records, fiber.StatusInternalServerError, errors.New("failed getting quick quotations")
	}

	fmt.Println("Quick Quotes: ", quotations)

	for _, v := range quotations {
		var quickquotations models.SalesQuotationQuick

		conditions := map[string]interface{}{
			"based_id": v.ID,
		}

		//CHILD 1
		if err := GetSalesQuotationQuick(&quickquotations, conditions); err != nil {
			return records, fiber.StatusInternalServerError, err
		}

		body := Body{
			SalesQuotation: v,
			QuickQuote:     quickquotations,
		}

		records = append(records, body)
	}

	return records, 0, nil
}

func GetSalesQuotation(id int) (Body, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var record Body

	if err := services.DbGet(&record.SalesQuotation, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}

	conditions = map[string]interface{}{
		// based on parent ID
		"based_id": record.SalesQuotation.ID,
	}

	//Child 1
	if err := GetSalesQuotationQuick(&record.QuickQuote, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

// CREATE SERVICE
func CreateSalesQuotation(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesQuotation); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SalesQuotationAt{RefId: body.ID, SalesQuotationContent: body.SalesQuotationContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	if err := CreateSalesQuotationQuick(tx, body.ID, body.QuickQuote, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	return body, 0, nil
}

func UpdateSalesQuotation(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Parent class
	if err := services.DbUpdate(tx, &body.SalesQuotation, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SalesQuotationAt{RefId: body.ID, SalesQuotationContent: body.SalesQuotationContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	if err := UpdateItemSpecs(tx, body.QuickQuote, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func DeleteSalesQuotation(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.SalesQuotation, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SalesQuotationAt{RefId: body.ID, SalesQuotationContent: body.SalesQuotationContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := DeleteItemSpecs(tx, body.QuickQuote, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
