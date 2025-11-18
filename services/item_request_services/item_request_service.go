package item_request_services

import (
	// "errors"

	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type ItemRequestBody struct {
	ItemRequest         models.ItemRequest           `json:"item_request"`
	ItemRequestDetails  []models.ItemRequestDetails  `json:"item_request_details"`
	ItemRequestLocation []models.ItemRequestLocation `json:"item_request_location"`
	ItemRequestHistory  []models.ItemRequestHistory  `json:"item_request_history"`
}

type ItemRequestGet struct {
	ItemRequest         []models.ItemRequest         `json:"item_request"`
	ItemRequestDetails  []models.ItemRequestDetails  `json:"item_request_details"`
	ItemRequestLocation []models.ItemRequestLocation `json:"item_request_location"`
}

func GetItemRequest(conditions map[string]interface{}) (interface{}, int, error) {
	var response ItemRequestGet

	if err := services.DbGet(&response.ItemRequest, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request")
	}

	if err := services.DbGet(&response.ItemRequestDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request details")
	}

	if err := services.DbGet(&response.ItemRequestLocation, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request locations")
	}

	return response, 0, nil
}

func GetAllItemList(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.AllItemView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting all item list")
	}

	return response, 0, nil
}

func GetAllBinLocation(itemId int64) (interface{}, int, error) {

	conditions := map[string]interface{}{
		"ItemId": itemId,
	}

	var response []models.AllBinLocationView

	if err := services.DbRaw(&response, "sp_GetBinLocationItem", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting all bin location")
	}

	return response, 0, nil
}

func GetUserList(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.UserListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting user list")
	}

	return response, 0, nil
}

func GetSalesOrderIR(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.SalesOrderViewIR

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting all sales order IR")
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

	generatedDocNo := utils.DocNoGenerator(body.ItemRequest.ID)
	body.ItemRequest.DocNo = generatedDocNo

	if err := tx.Model(&body.ItemRequest).Update("doc_no", body.ItemRequest.DocNo).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating item request doc")
	}

	//Prepare the "at" data
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Insert Item Request Details
	if err := CreateItemRequestDetails(tx, &body, at); err != nil {
		return body.ItemRequest, fiber.StatusInternalServerError, err
	}

	if err := CreateItemRequestHistory(tx, &body, at); err != nil {
		return body.ItemRequest, fiber.StatusInternalServerError, err
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

func CreateItemRequestDetails(tx *gorm.DB, body *ItemRequestBody, at models.At) error {
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]
		detail.IrId = body.ItemRequest.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating item request details")
		}

		// Audit trail for each detail
		atdataDetail := models.ItemRequestDetailsAt{
			RefId:                     detail.ID,
			ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating item request details at")
		}
	}
	return nil
}

func CreateItemRequestLocations(tx *gorm.DB, body *ItemRequestBody, at models.At) error {
	for i := range body.ItemRequestLocation {
		location := &body.ItemRequestLocation[i]
		location.IrId = body.ItemRequest.ID // assign FK to parent

		inventory := models.InventoryStocksHistory{
			InventoryStocksHistoryContent: models.InventoryStocksHistoryContent{
				ItemRequestId:        location.IrId,
				ItemRequestDetailsId: location.IrDetailsId,
				StockQty:             location.StockQty,
				ReqQty:               location.IssuedQty,
				ItemId:               location.ItemId,
				BinLocation:          location.Location,
			},
		}

		if err := services.DbInsert(tx, location); err != nil {
			return errors.New("failed creating item request location")
		}

		if err := setup_services.CreateInventoryStocksHistory(tx, &inventory, at); err != nil {
			return err
		}

		// Audit trail for each location
		atdataLocation := models.ItemRequestLocationAt{
			RefId:                      location.ID,
			ItemRequestLocationContent: location.ItemRequestLocationContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataLocation); err != nil {
			return errors.New("failed creating item request location at")
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}

func CreateItemRequestHistory(tx *gorm.DB, body *ItemRequestBody, at models.At) error {
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]

		// Create a history entry for each item request detail
		history := models.ItemRequestHistory{
			ItemRequestHistoryContent: models.ItemRequestHistoryContent{
				IRId:     body.ItemRequest.ID,
				IRDId:    detail.ID,
				RefDoc:   body.ItemRequest.RefDoc,
				ItemID:   detail.ItemId,
				ReqDate:  body.ItemRequest.ReqDate,
				OrderQty: detail.OrderQty,
				ReqQty:   detail.ReqQty,
				SOId:     detail.SOId,
				SODId:    detail.SODId,
			},
		}

		// Set current date in MM/dd/yyyy format
		history.TransactionDate = time.Now().Format("01/02/2006")

		if err := services.DbInsert(tx, &history); err != nil {
			return errors.New("failed creating item request history")
		}

		// Audit trail for each history record
		atdataHistory := models.ItemRequestHistoryAt{
			RefId:                     history.ID,
			ItemRequestHistoryContent: history.ItemRequestHistoryContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataHistory); err != nil {
			return errors.New("failed creating item request history at")
		}
	}

	setup_services.InvalidateItemCaches()
	return nil
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

	// Handle details
	if err := UpdateItemRequestDetails(tx, &body, conditions, at); err != nil {
		return body.ItemRequest, fiber.StatusInternalServerError, err
	}

	// Handle locations
	if err := UpdateItemRequestLocations(tx, &body, conditions, at); err != nil {
		return body.ItemRequest, fiber.StatusInternalServerError, err
	}

	// Inside Update Item Request (after updating details and locations)
	if err := UpdateItemRequestHistory(tx, &body, conditions, at); err != nil {
		return body.ItemRequest, fiber.StatusInternalServerError, err
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

func UpdateItemRequestDetails(tx *gorm.DB, body *ItemRequestBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]
		detail.IrId = body.ItemRequest.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating item request details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating item request details")
			}
		}

		// Audit record for each detail
		atdataDetail := models.ItemRequestDetailsAt{
			RefId:                     detail.ID,
			ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating item request details at")
		}
	}
	return nil
}

func UpdateItemRequestLocations(tx *gorm.DB, body *ItemRequestBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.ItemRequestLocation {
		location := &body.ItemRequestLocation[i]
		location.IrId = body.ItemRequest.ID

		inventory := models.InventoryStocksHistory{
			InventoryStocksHistoryContent: models.InventoryStocksHistoryContent{
				ItemRequestId:        location.IrId,
				ItemRequestDetailsId: location.IrDetailsId,
				ItemId:               location.ItemId,
				BinLocation:          location.Location,
				StockQty:             location.StockQty,
				ReqQty:               location.IssuedQty,
			},
		}

		if location.ID == 0 {
			if err := services.DbInsert(tx, location); err != nil {
				return errors.New("failed creating item request location")
			}
		} else {
			if err := services.DbUpdate(tx, location, conditions); err != nil {
				return errors.New("failed updating item request location")
			}
		}

		if err := setup_services.UpdateInventoryStocksHistory(tx, &inventory, at); err != nil {
			return err
		}

		// Audit record for each location
		atdataLocation := models.ItemRequestLocationAt{
			RefId:                      location.ID,
			ItemRequestLocationContent: location.ItemRequestLocationContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataLocation); err != nil {
			return errors.New("failed creating item request location at")
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}

func UpdateItemRequestHistory(tx *gorm.DB, body *ItemRequestBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.ItemRequestDetails {
		detail := &body.ItemRequestDetails[i]

		//Skip if SOId or SODId are empty (zero)
		if detail.SOId == 0 || detail.SODId == 0 {
			continue
		}

		// Build or update the history record
		history := models.ItemRequestHistory{
			ItemRequestHistoryContent: models.ItemRequestHistoryContent{
				IRId:     body.ItemRequest.ID,
				IRDId:    detail.ID,
				ItemID:   detail.ItemId,
				ReqDate:  body.ItemRequest.ReqDate,
				OrderQty: detail.OrderQty,
				ReqQty:   detail.ReqQty,
				SOId:     detail.SOId,
				SODId:    detail.SODId,
			},
		}

		// Set current date in MM/dd/yyyy format
		history.TransactionDate = time.Now().Format("01/02/2006")

		// If an existing record exists (IRDId + IRId combination), update it
		var existing models.ItemRequestHistory
		err := tx.Where("ir_id = ? AND ird_id = ?", body.ItemRequest.ID, detail.ID).First(&existing).Error
		if err == nil {
			history.ID = existing.ID // ensure update, not insert
			if err := services.DbUpdate(tx, &history, map[string]interface{}{"id": existing.ID}); err != nil {
				return errors.New("failed updating item request history")
			}
		} else {
			return errors.New("failed fetching item request history for update")
		}

		// Insert audit trail record for history update
		atdataHistory := models.ItemRequestHistoryAt{
			RefId:                     history.ID,
			ItemRequestHistoryContent: history.ItemRequestHistoryContent,
			At:                        at,
		}

		if err := services.DbInsert(tx, &atdataHistory); err != nil {
			return errors.New("failed creating item request history audit record")
		}
	}

	return nil
}

func DeleteItemRequest(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body ItemRequestBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	//Delete main Item Request
	if err := services.DbDelete(tx, &body.ItemRequest, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item request")
	}

	if err := DeleteItemRequestDetails(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeleteItemRequestLocations(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeleteItemRequestHistory(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
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

	// Optionally fetch deleted location to log in audit trail
	var deletedLocation []models.ItemRequestLocation
	if err := tx.Unscoped().Where("ir_id = ?", body.ItemRequest.ID).Find(&deletedLocation).Error; err == nil {
		for _, location := range deletedLocation {
			atdataLocation := models.ItemRequestLocationAt{
				RefId:                      location.ID,
				ItemRequestLocationContent: location.ItemRequestLocationContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataLocation); err != nil {
				return body.ItemRequest, fiber.StatusInternalServerError, errors.New("failed creating item request location audit record")
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

func DeleteItemRequestDetails(tx *gorm.DB, body *ItemRequestBody, at models.At) error {
	// Delete all details
	if err := services.DbDelete(tx, &models.ItemRequestDetails{}, map[string]interface{}{"ir_id": body.ItemRequest.ID}); err != nil {
		return errors.New("failed deleting all item request details")
	}

	// Optionally fetch deleted details for audit trail
	var deletedDetails []models.ItemRequestDetails
	if err := tx.Unscoped().Where("ir_id = ?", body.ItemRequest.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := models.ItemRequestDetailsAt{
				RefId:                     detail.ID,
				ItemRequestDetailsContent: detail.ItemRequestDetailsContent,
				At:                        at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating item request details audit record")
			}
		}
	}
	return nil
}

func DeleteItemRequestLocations(tx *gorm.DB, body *ItemRequestBody, at models.At) error {
	// Delete all locations
	if err := services.DbDelete(tx, &models.ItemRequestLocation{}, map[string]interface{}{"ir_id": body.ItemRequest.ID}); err != nil {
		return errors.New("failed deleting all item request location")
	}

	// Optionally fetch deleted locations for audit trail
	var deletedLocations []models.ItemRequestLocation
	if err := tx.Unscoped().Where("ir_id = ?", body.ItemRequest.ID).Find(&deletedLocations).Error; err == nil {
		for _, location := range deletedLocations {
			atdataLocation := models.ItemRequestLocationAt{
				RefId:                      location.ID,
				ItemRequestLocationContent: location.ItemRequestLocationContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataLocation); err != nil {
				return errors.New("failed creating item request location audit record")
			}
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}

func DeleteItemRequestHistory(tx *gorm.DB, body *ItemRequestBody, at models.At) error {
	// Delete all histories linked to the Item Request
	if err := services.DbDelete(tx, &models.ItemRequestHistory{}, map[string]interface{}{"ir_id": body.ItemRequest.ID}); err != nil {
		return errors.New("failed deleting all item request history")
	}

	inventory := models.InventoryStocksHistory{
		InventoryStocksHistoryContent: models.InventoryStocksHistoryContent{
			ItemRequestId: body.ItemRequest.ID,
		},
	}

	if err := setup_services.DeleteInventoryStocksHistory(tx, &inventory, at); err != nil {
		return err
	}

	// Optionally fetch deleted history records (Unscoped for audit)
	var deletedHistories []models.ItemRequestHistory
	if err := tx.Unscoped().Where("ir_id = ?", body.ItemRequest.ID).Find(&deletedHistories).Error; err == nil {
		for _, history := range deletedHistories {
			atdataHistory := models.ItemRequestHistoryAt{
				RefId:                     history.ID,
				ItemRequestHistoryContent: history.ItemRequestHistoryContent,
				At:                        at,
			}
			if err := services.DbInsert(tx, &atdataHistory); err != nil {
				return errors.New("failed creating item request history audit record")
			}
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}
