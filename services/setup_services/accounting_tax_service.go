package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type TaxSetupBody struct {
	TaxSetup        accounting_models.Tax          `json:"tax_setup"`
	TaxSetupDetails []accounting_models.TaxDetails `json:"tax_setup_details"`
}

type TaxSetupGet struct {
	TaxSetup        []accounting_models.Tax            `json:"tax_setup"`
	TaxSetupDetails []accounting_models.TaxDetailsView `json:"tax_setup_details"`
	TaxSetupView    []accounting_models.TaxView        `json:"tax_setup_view"`
}

func GetTaxSetup(conditions map[string]interface{}) (interface{}, int, error) {

	var response TaxSetupGet

	if err := services.DbGet(&response.TaxSetup, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting tax setup")
	}

	if err := services.DbGet(&response.TaxSetupDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting tax setup details")
	}

	if err := services.DbGet(&response.TaxSetupView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting tax setup view")
	}

	return response, 0, nil
}

func GetTaxClassificationSetup(code string) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"code": code,
	}
	var response []accounting_models.TaxClassification

	if err := services.DbRaw(&response, "sp_tax_setup_classification", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get tax classification setup")
	}

	return response, 0, nil
}

func GetChartOfAccountSetup(conditions map[string]interface{}) (interface{}, int, error) {

	var response []accounting_models.CoaView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting chart of accounts view")
	}

	return response, 0, nil
}

func CreateTaxSetup(c *fiber.Ctx, tx *gorm.DB) (interface{}, int, error) {
	var body TaxSetupBody

	//Parse the full request body (main + details)
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Insert main Tax Setup record
	if err := services.DbInsert(tx, &body.TaxSetup); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating tax setup")
	}

	//Prepare the "at" data
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Insert Tax Details
	if err := CreateTaxDetails(tx, &body, at); err != nil {
		return body.TaxSetup, fiber.StatusInternalServerError, err
	}

	//Insert audit record for the main request
	atdata := accounting_models.TaxAt{
		RefId:      body.TaxSetup.ID,
		TaxContent: body.TaxSetup.TaxContent,
		At:         at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating tax setup at")
	}

	return body, 0, nil
}

func CreateTaxDetails(tx *gorm.DB, body *TaxSetupBody, at models.At) error {
	for i := range body.TaxSetupDetails {
		detail := &body.TaxSetupDetails[i]
		detail.TaxCodeId = body.TaxSetup.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating tax details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.TaxDetailsAt{
			RefId:             detail.ID,
			TaxDetailsContent: detail.TaxDetailsContent,
			At:                at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating tax details at")
		}
	}

	InvalidateItemCaches()

	return nil
}

func UpdateTaxSetup(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body TaxSetupBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Update main Tax Setup
	if err := services.DbUpdate(tx, &body.TaxSetup, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating tax setup")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Handle details
	if err := UpdateTaxDetails(tx, &body, conditions, at); err != nil {
		return body.TaxSetup, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := accounting_models.TaxAt{
		RefId:      body.TaxSetup.ID,
		TaxContent: body.TaxSetup.TaxContent,
		At:         at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating tax setup at")
	}

	return body, 0, nil
}

func UpdateTaxDetails(tx *gorm.DB, body *TaxSetupBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.TaxSetupDetails {
		detail := &body.TaxSetupDetails[i]
		detail.TaxCodeId = body.TaxSetup.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating tax details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating tax details")
			}
		}

		// Audit record for each detail
		atdataDetail := accounting_models.TaxDetailsAt{
			RefId:             detail.ID,
			TaxDetailsContent: detail.TaxDetailsContent,
			At:                at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating tax details at")
		}
	}

	InvalidateItemCaches()

	return nil
}

func DeleteTaxSetup(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body TaxSetupBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	//Delete main Tax Setup
	if err := services.DbDelete(tx, &body.TaxSetup, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting tax setup")
	}

	if err := DeleteTaxDetails(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := accounting_models.TaxAt{RefId: body.TaxSetup.ID, TaxContent: body.TaxSetup.TaxContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating tax setup at")
	}

	return body, 0, nil
}

func DeleteTaxDetails(tx *gorm.DB, body *TaxSetupBody, at models.At) error {
	// Delete all details
	if err := services.DbDelete(tx, &accounting_models.TaxDetails{}, map[string]interface{}{"tax_code_id": body.TaxSetup.ID}); err != nil {
		return errors.New("failed deleting all tax details")
	}

	// Optionally fetch deleted details for audit trail
	var deletedDetails []accounting_models.TaxDetails
	if err := tx.Unscoped().Where("tax_code_id = ?", body.TaxSetup.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := accounting_models.TaxDetailsAt{
				RefId:             detail.ID,
				TaxDetailsContent: detail.TaxDetailsContent,
				At:                at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating tax details audit record")
			}
		}
	}

	InvalidateItemCaches()

	return nil
}
