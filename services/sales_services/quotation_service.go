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
	//Child 1
	SalesQuotationQuick models.SalesQuotationQuick `json:"sales_quotation_quick"`
}

func GetSalesQuotations(conditions map[string]interface{}) (interface{}, int, error) {

	type Response struct {
		SalesQuotation      []models.SalesQuotation
		SalesQuotationQuick []models.SalesQuotationQuick
	}

	var response Response

	if err := services.DbGet(&response.SalesQuotation, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales quotations")
	}

	if err := GetSalesQuotationQuicks(&response.SalesQuotationQuick, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
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
	if err := GetSalesQuotationQuick(&record.SalesQuotationQuick, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	// //Child 2 for project quote
	// if err := GetSalesQuotationQuick(&record.QuickQuote, conditions); err != nil {
	// 	return record, fiber.StatusInternalServerError, err
	// }

	return record, 0, nil
}

// CREATE CHILD SERVICE
func CreateSalesQuotationChild(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := CreateSalesQuotationQuick(tx, body.SalesQuotationQuick, at); err != nil {
		fmt.Println("err", err)
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

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
	return body, 0, nil
}

func UpdateSalesQuotation(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		fmt.Println(err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Parent class
	if err := services.DbUpdate(tx, &body.SalesQuotation, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating quotation")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SalesQuotationAt{RefId: body.ID, SalesQuotationContent: body.SalesQuotationContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating quotationat")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := UpdateSalesQuotationQuick(tx, body.SalesQuotationQuick, at, conditions); err != nil {
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

	if err := DeleteSalesQuotationQuick(tx, body.SalesQuotationQuick, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
