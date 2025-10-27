package setup_services

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type ItemRequestBody struct {
	ItemRequest        models.ItemRequest          `json:"item_request"`
	ItemRequestDetails []models.ItemRequestDetails `json:"item_request_details"`
}

type ItemRequestGet struct {
	ItemRequest        []models.ItemRequest        `json:"item_request"`
	ItemRequestDetails []models.ItemRequestDetails `json:"item_request_details"`
}

func GetItemRequest(conditions map[string]interface{}) (interface{}, int, error) {
	var response ItemRequestGet

	if err := services.DbGet(&response.ItemRequest, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request")
	}

	if err := services.DbGet(&response.ItemRequestDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request details")
	}

	return response, 0, nil
}

func CreateItemRequest(c *fiber.Ctx, tx *gorm.DB) (interface{}, int, error) {
	var body ItemRequestBody

	//Parse the full request body (main + details)
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Insert main Item Request record
	if err := services.DbInsert(tx, &body.ItemRequest); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item request")
	}

	//Prepare the "at" data
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	//Insert Item Request Details (if any)
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]
		detail.IrId = body.ItemRequest.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed creating item request details")
		}

		// Insert audit record for each detail
		atdataDetail := models.ItemRequestDetailsAt{
			RefId:                     detail.ID, // the ID of the inserted detail
			ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed creating item request details at")
		}
	}

	//Insert audit record for the main request
	atdata := models.ItemRequestAt{
		RefId:              body.ItemRequest.ID,
		ItemRequestContent: body.ItemRequest.ItemRequestContent,
		At:                 at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item request at")
	}

	return body, 0, nil
}

func UpdateItemRequest(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body ItemRequestBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Update main Item Request
	if err := services.DbUpdate(tx, &body.ItemRequest, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item request")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	//Update details
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]
		detail.IrId = body.ItemRequest.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed creating item request details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed updating item request details")
			}
		}

		// Audit record for each detail
		atdataDetail := models.ItemRequestDetailsAt{
			RefId:                     detail.ID,
			ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed creating item request details at")
		}
	}

	//Audit record for main request
	atdata := models.ItemRequestAt{
		RefId:              body.ItemRequest.ID,
		ItemRequestContent: body.ItemRequest.ItemRequestContent,
		At:                 at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item request at")
	}

	return body, 0, nil
}

func DeleteItemRequest(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body ItemRequestBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Delete main Item Request
	if err := services.DbDelete(tx, &body.ItemRequest, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item request")
	}

	// Delete all details where ir_id == body.ItemRequest.ID
	if err := services.DbDelete(tx, &models.ItemRequestDetails{}, map[string]interface{}{"ir_id": body.ItemRequest.ID}); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting all item request details")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Optionally fetch deleted details to log in audit trail
	var deletedDetails []models.ItemRequestDetails
	if err := tx.Unscoped().Where("ir_id = ?", body.ItemRequest.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := models.ItemRequestDetailsAt{
				RefId:                     detail.ID,
				ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
				At:                        at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed creating item request details audit record")
			}
		}
	}

	//Audit record for main request
	atdata := models.ItemRequestAt{RefId: body.ItemRequest.ID, ItemRequestContent: body.ItemRequest.ItemRequestContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating item request at")
	}

	return body, 0, nil
}
