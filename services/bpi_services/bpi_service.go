package bpi_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type Body struct {
	models.Bpi
	IndustriesId []uint                  `json:"industries_id"`
	General      models.BpiGeneralSchema `json:"general"`
	Contacts     []models.BpiContacts    `json:"contacts"`
	Address      []models.BpiAddress     `json:"address"`
	Items        []models.BpiItems       `json:"items"`
}

type BodySample struct {
	models.SetupItemBom
	BomDetails []models.SetupItemBomDetails `json:"bom_details"`
}

func GetBpis(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Bpi      []models.BpiView        `json:"bpi"`
		General  []models.BpiGeneralView `json:"general"`
		Contacts []models.BpiContactView `json:"contacts"`
		Address  []models.BpiAddressView `json:"address"`
		Items    []models.BpiItemsView   `json:"items"`
	}

	var response Response

	if err := services.DbGet(&response.Bpi, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpis")
	}

	if err := services.DbGet(&response.General, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi general list")
	}

	if err := services.DbGet(&response.Contacts, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi contacts")
	}

	if err := services.DbGet(&response.Address, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi address")
	}
	if err := services.DbGet(&response.Items, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi items")
	}

	return response, 0, nil
}

func GetBpiItemList(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.BpiItemList

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi item list")
	}

	return response, 0, nil
}

func CreateBpi(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.Bpi); err != nil {
		return body, fiber.StatusInternalServerError, errors.New(("failed creating parent"))
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	parentat := models.BpiAt{RefId: body.ID, BpiContent: models.BpiContent{SalesId: body.SalesId, Name: body.Name, Tin: body.Tin, MainTelNo: body.MainTelNo}, At: at}
	if err := services.DbInsert(tx, &parentat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	for _, v := range body.IndustriesId {
		if err := CreateBpiIndustries(tx, body.ID, uint(v), at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Create Bpi General

	if err := BpiGeneral(tx, body.ID, body.General, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Create  Bpi Contacts

	for _, v := range body.Contacts {
		fmt.Println("Branch Contacts", v)
		if err := BpiContact(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Create Bpi Address

	for _, v := range body.Address {

		if err := BpiAddress(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Create Bpi Items
	if err := BpiItems(tx, body.ID, body.Items, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
