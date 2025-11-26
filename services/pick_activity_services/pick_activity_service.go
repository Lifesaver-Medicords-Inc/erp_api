package pick_activity_services

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

type PickActivityBody struct {
	PickActivity         models.PickActivity           `json:"pick_activity"`
	PickActivityDetails  []models.PickActivityDetails  `json:"pick_activity_details"`
	PickActivityLocation []models.PickActivityLocation `json:"pick_activity_location"`
	PickActivityHistory  []models.PickActivityHistory  `json:"pick_activity_history"`
}

type PickActivityGet struct {
	PickActivity         []models.PickActivity         `json:"pick_activity"`
	PickActivityDetails  []models.PickActivityDetails  `json:"pick_activity_details"`
	PickActivityLocation []models.PickActivityLocation `json:"pick_activity_location"`
}

func GetPickActivity(conditions map[string]interface{}) (interface{}, int, error) {
	var response PickActivityGet

	if err := services.DbGet(&response.PickActivity, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity")
	}

	if err := services.DbGet(&response.PickActivityDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity details")
	}

	if err := services.DbGet(&response.PickActivityLocation, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity locations")
	}

	return response, 0, nil
}

func GetBinLocation(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.WarehouseArea

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting all bin location PA")
	}

	return response, 0, nil
}

func GetSalesOrderPA(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.SalesOrderViewPA

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting all sales order PA")
	}

	return response, 0, nil
}

func CreatePickActivity(c *fiber.Ctx, tx *gorm.DB) (interface{}, int, error) {
	var body PickActivityBody

	//Parse the full request body (main + details)
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Set current date in MM/dd/yyyy format
	body.PickActivity.TransactionDate = time.Now().Format("01/02/2006")

	//Insert main Pick Activity record
	if err := services.DbInsert(tx, &body.PickActivity); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity")
	}

	generatedDocNo := utils.DocNoGenerator(body.PickActivity.ID)
	body.PickActivity.DocNo = generatedDocNo

	if err := tx.Model(&body.PickActivity).Update("doc_no", body.PickActivity.DocNo).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity doc")
	}

	//Prepare the "at" data
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Insert Pick Activity Details
	if err := CreatePickActivityDetails(tx, &body, at); err != nil {
		return body.PickActivity, fiber.StatusInternalServerError, err
	}

	//Only create Pick Activity History if RefDoc is not empty
	if body.PickActivity.ReferenceSo != "" {
		if err := CreatePickActivityHistory(tx, &body, at); err != nil {
			return body.PickActivity, fiber.StatusInternalServerError, err
		}
	}

	//Insert audit record for the main request
	atdata := models.PickActivityAt{
		RefId:               body.PickActivity.ID,
		PickActivityContent: body.PickActivity.PickActivityContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity at")
	}

	return body, 0, nil
}

func CreatePickActivityDetails(tx *gorm.DB, body *PickActivityBody, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]
		detail.PaId = body.PickActivity.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating pick activity details")
		}

		// Audit trail for each detail
		atdataDetail := models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details at")
		}
	}
	return nil
}

func CreatePickActivityHistory(tx *gorm.DB, body *PickActivityBody, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]

		// Create a history entry for each pick activity detail
		history := models.PickActivityHistory{
			PickActivityHistoryContent: models.PickActivityHistoryContent{
				PAId:    body.PickActivity.ID,
				PADId:   detail.ID,
				RefDoc:  body.PickActivity.ReferenceSo,
				ItemID:  detail.ItemId,
				LeftQty: detail.LeftQty,
				PickQty: detail.PickQty,
				SOId:    detail.SOId,
				SODId:   detail.SODId,
			},
		}

		// Set current date in MM/dd/yyyy format
		history.TransactionDate = time.Now().Format("01/02/2006")

		if err := services.DbInsert(tx, &history); err != nil {
			return errors.New("failed creating pick activity history")
		}

		// Audit trail for each history record
		atdataHistory := models.PickActivityHistoryAt{
			RefId:                      history.ID,
			PickActivityHistoryContent: history.PickActivityHistoryContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataHistory); err != nil {
			return errors.New("failed creating pick activity history at")
		}
	}

	setup_services.InvalidateItemCaches()
	return nil
}

func UpdatePickActivity(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body PickActivityBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Set current date in MM/dd/yyyy format
	body.PickActivity.TransactionDate = time.Now().Format("01/02/2006")

	//Update main Pick Activity
	if err := services.DbUpdate(tx, &body.PickActivity, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Handle locations
	if err := UpdatePickActivityDetails(tx, &body, conditions, at); err != nil {
		return body.PickActivity, fiber.StatusInternalServerError, err
	}

	// Inside Update Pick Activity (after updating details and locations)
	if err := UpdatePickActivityHistory(tx, &body, conditions, at); err != nil {
		return body.PickActivity, fiber.StatusInternalServerError, err
	}

	// Inside Update Pick Activity (after updating details and locations)
	if err := UpdatePickActivityLocations(tx, &body, conditions, at); err != nil {
		return body.PickActivity, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := models.PickActivityAt{
		RefId:               body.PickActivity.ID,
		PickActivityContent: body.PickActivity.PickActivityContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity at")
	}

	return body, 0, nil
}

func UpdatePickActivityDetails(tx *gorm.DB, body *PickActivityBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]
		detail.PaId = body.PickActivity.ID
		TransDate := body.PickActivity.TransactionDate

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating pick activity details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating pick activity details")
			}
		}

		// Audit record for each detail
		atdataDetail := models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details at")
		}

		if detail.BinLocation == "" || detail.WarehouseId == 0 || detail.ActualQty == 0 {
			continue
		}

		// Prepare inventory stock struct
		inventory := models.InventoryStocks{
			InventoryStocksContent: models.InventoryStocksContent{
				PickActivityId:         detail.PaId,
				PickActivityDetailsId:  detail.ID,
				PurchaseOrderDetailsId: detail.SODId,
				ItemId:                 detail.ItemId,
				BinLocation:            detail.BinLocation,
				Uom:                    detail.ActualUom,
				QtyIn:                  detail.ActualQty,
				WarehouseId:            detail.WarehouseId,
				SupplierName:           body.PickActivity.Customer,
				DateReceived:           TransDate,
			},
		}

		var existing models.InventoryStocks

		err := tx.Where("item_id = ? AND pick_activity_id = ? AND pick_activity_details_id = ?", inventory.ItemId, inventory.PickActivityId, inventory.PickActivityDetailsId).First(&existing).Error

		if err == nil && existing.ReceivingReportDetailsId == 0 && existing.ReceivingReportId == 0 {
			if err := setup_services.UpdateInventoryStockPickActivity(tx, &inventory, at); err != nil {
				return err
			}
		} else {
			if err := setup_services.CreateInventoryStock(tx, &inventory, at); err != nil {
				return err
			}
		}
	}
	return nil
}

func UpdatePickActivityLocations(tx *gorm.DB, body *PickActivityBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.PickActivityLocation {
		location := &body.PickActivityLocation[i]
		location.PaId = body.PickActivity.ID

		inventory := models.InventoryStocksHistory{
			InventoryStocksHistoryContent: models.InventoryStocksHistoryContent{
				PickActivityId:        location.PaId,
				PickActivityDetailsId: location.PaDetailsId,
				ItemId:                location.ItemId,
				BinLocation:           location.Location,
				StockQty:              location.StockQty,
				ReqQty:                location.ActualQty,
				WarehouseId:           location.WarehouseId,
			},
		}

		if location.ID == 0 {
			if err := services.DbInsert(tx, location); err != nil {
				return errors.New("failed creating pick activity location")
			}
		} else {
			if err := services.DbUpdate(tx, location, conditions); err != nil {
				return errors.New("failed updating pick activity location")
			}
		}

		if err := setup_services.UpdateInventoryStocksHistory(tx, &inventory, at); err != nil {
			return err
		}

		// Audit record for each location
		atdataLocation := models.PickActivityLocationAt{
			RefId:                       location.ID,
			PickActivityLocationContent: location.PickActivityLocationContent,
			At:                          at,
		}

		if err := services.DbInsert(tx, &atdataLocation); err != nil {
			return errors.New("failed creating pick activity location at")
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}

func UpdatePickActivityHistory(tx *gorm.DB, body *PickActivityBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]

		// Build or update the history record
		history := models.PickActivityHistory{
			PickActivityHistoryContent: models.PickActivityHistoryContent{
				PAId:    body.PickActivity.ID,
				PADId:   detail.ID,
				ItemID:  detail.ItemId,
				LeftQty: detail.LeftQty,
				PickQty: detail.PickQty,
				SOId:    detail.SOId,
				SODId:   detail.SODId,
			},
		}

		// Set current date in MM/dd/yyyy format
		history.TransactionDate = time.Now().Format("01/02/2006")

		// If an existing record exists (IRDId + IRId combination), update it
		var existing models.PickActivityHistory
		err := tx.Where("pa_id = ? AND pad_id = ?", body.PickActivity.ID, detail.ID).First(&existing).Error
		if err == nil {
			history.ID = existing.ID // ensure update, not insert
			if err := services.DbUpdate(tx, &history, map[string]interface{}{"id": existing.ID}); err != nil {
				return errors.New("failed updating pick activity history")
			}
		} else {
			return errors.New("failed fetching pick activity history for update")
		}

		// Insert audit trail record for history update
		atdataHistory := models.PickActivityHistoryAt{
			RefId:                      history.ID,
			PickActivityHistoryContent: history.PickActivityHistoryContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataHistory); err != nil {
			return errors.New("failed creating pick activity history audit record")
		}
	}

	return nil
}

func DeletePickActivity(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body PickActivityBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	//Delete main Pick Activity
	if err := services.DbDelete(tx, &body.PickActivity, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting pick activity")
	}

	if err := DeletePickActivityDetails(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeletePickActivityLocations(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeletePickActivityHistory(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Optionally fetch deleted details to log in audit trail
	var deletedDetails []models.PickActivityDetails
	if err := tx.Unscoped().Where("pa_id = ?", body.PickActivity.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := models.PickActivityDetailsAt{
				RefId:                      detail.ID,
				PickActivityDetailsContent: detail.PickActivityDetailsContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return body.PickActivity, fiber.StatusInternalServerError, errors.New("failed creating pick activity details audit record")
			}
		}
	}

	//Audit record for main request
	atdata := models.PickActivityAt{RefId: body.PickActivity.ID, PickActivityContent: body.PickActivity.PickActivityContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity at")
	}

	return body, 0, nil
}

func DeletePickActivityDetails(tx *gorm.DB, body *PickActivityBody, at models.At) error {
	// Delete all details
	if err := services.DbDelete(tx, &models.PickActivityDetails{}, map[string]interface{}{"pa_id": body.PickActivity.ID}); err != nil {
		return errors.New("failed deleting all pick activity details")
	}

	// Optionally fetch deleted details for audit trail
	var deletedDetails []models.PickActivityDetails
	if err := tx.Unscoped().Where("pa_id = ?", body.PickActivity.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := models.PickActivityDetailsAt{
				RefId:                      detail.ID,
				PickActivityDetailsContent: detail.PickActivityDetailsContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating pick activity details audit record")
			}
		}
	}
	return nil
}

func DeletePickActivityLocations(tx *gorm.DB, body *PickActivityBody, at models.At) error {
	// Delete all locations
	if err := services.DbDelete(tx, &models.PickActivityLocation{}, map[string]interface{}{"pa_id": body.PickActivity.ID}); err != nil {
		return errors.New("failed deleting all pick activity location")
	}

	inventory := models.InventoryStocksHistory{
		InventoryStocksHistoryContent: models.InventoryStocksHistoryContent{
			PickActivityId: body.PickActivity.ID,
		},
	}

	if err := setup_services.DeleteInventoryStocksPAHistory(tx, &inventory, at); err != nil {
		return err
	}

	if err := setup_services.DeleteInventoryStock(tx, body.PickActivity.ID, 0, at); err != nil {
		return err
	}

	// Optionally fetch deleted locations for audit trail
	var deletedLocations []models.PickActivityLocation
	if err := tx.Unscoped().Where("pa_id = ?", body.PickActivity.ID).Find(&deletedLocations).Error; err == nil {
		for _, location := range deletedLocations {
			atdataLocation := models.PickActivityLocationAt{
				RefId:                       location.ID,
				PickActivityLocationContent: location.PickActivityLocationContent,
				At:                          at,
			}
			if err := services.DbInsert(tx, &atdataLocation); err != nil {
				return errors.New("failed creating pick activity location audit record")
			}
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}

func DeletePickActivityHistory(tx *gorm.DB, body *PickActivityBody, at models.At) error {
	// Delete all histories linked to the Pick Activity
	if err := services.DbDelete(tx, &models.PickActivityHistory{}, map[string]interface{}{"pa_id": body.PickActivity.ID}); err != nil {
		return errors.New("failed deleting all pick activity history")
	}

	// Optionally fetch deleted history records (Unscoped for audit)
	var deletedHistories []models.PickActivityHistory
	if err := tx.Unscoped().Where("pa_id = ?", body.PickActivity.ID).Find(&deletedHistories).Error; err == nil {
		for _, history := range deletedHistories {
			atdataHistory := models.PickActivityHistoryAt{
				RefId:                      history.ID,
				PickActivityHistoryContent: history.PickActivityHistoryContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataHistory); err != nil {
				return errors.New("failed creating pick activity history audit record")
			}
		}
	}

	setup_services.InvalidateItemCaches()

	return nil
}
