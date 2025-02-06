package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type CustomerBody struct {
	models.Bpi
	General  models.BpiGeneral  `json:"general"`
	Contacts models.BpiContacts `json:"contacts"`
	Address  models.BpiAddress  `json:"address"`
}

type Body struct {
	models.SalesQuotation
	//Child 1
	SalesQuotationQuick models.SalesQuotationQuick `json:"sales_quotation_quick"`
}

type CreateBody struct {
	models.SalesQuotation
	//Child 1
	SalesQuotationQuick []models.SalesQuotationQuick `json:"sales_quotation_quick"`
}

type ItemBody struct {
	Items    models.Item `json:"items"`
	ItemName models.Name `json:"item_name"`
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

	return record, 0, nil
}

func CreateSalesQuotation(c *fiber.Ctx, tx *gorm.DB) (CreateBody, int, error) {
	var body CreateBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("pt 1", err)
		fmt.Println("ERR", body)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesQuotation); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", body)
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

	for _, v := range body.SalesQuotationQuick {
		if err := CreateSalesQuotationQuick(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// if err := CreateSalesQuotationQuick(tx, body.ID, body.SalesQuotationQuick[], at); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

func GetBpis(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		GetBpiCustomer []models.BpiCustomerView
	}

	var response Response

	if err := services.DbGet(&response.GetBpiCustomer, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi customer list")
	}

	return response, 0, nil
}

func GetBpi(id int) (CustomerBody, int, error) {

	conditions := map[string]interface{}{
		"id": id,
	}

	var record CustomerBody

	if err := services.DbGet(&record.Bpi, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}

	conditions = map[string]interface{}{
		// based on parent ID
		"based_id": record.Bpi.ID,
	}

	//Child 1
	if err := services.DbGet(&record.General, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := services.DbGet(&record.Address, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := services.DbGet(&record.Contacts, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}
	fmt.Println("DATA", record)
	return record, 0, nil
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items           []models.Item            `json:"items"`
		ItemSpecs       []models.ItemSpecs       `json:"itemspecs"`
		AdditionalSpecs []models.AdditionalSpecs `json:"additionalspecs"`
	}

	var response Response

	if err := services.DbGet(&response.Items, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting items")
	}

	fmt.Println(response)
	return response, 0, nil
}

// GET: item name based on item_name_id
func GetItem(id int) (ItemBody, int, error) {

	conditions := map[string]interface{}{
		"item_name_id": id,
	}

	var record ItemBody

	if err := services.DbGet(&record.Items, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}

	conditions = map[string]interface{}{
		"id": id,
	}

	if err := services.DbGet(&record.ItemName, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	//fmt.Println("DATAA", record)

	return record, 0, nil
}
