package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type DeleteReceivingReportBody struct {
	ReceivingReport          models.ReceivingReport          `json:"receiving_report"`
	ReceivingReportDetails   models.ReceivingReportDetails   `json:"receiving_report_details"` //child1
	ReceivingReportInventory models.ReceivingReportInventory `json:"receiving_report_inventory"`
	//child3 ReceivingReportAttachment models.ReceivingReportAttachment `json:"receiving_report_attachment
}

type ReceivingReportBody struct {
	ReceivingReport          models.ReceivingReport            `json:"receiving_report"`
	ReceivingReportDetails   []models.ReceivingReportDetails   `json:"receiving_report_details"` //child1
	ReceivingReportInventory []models.ReceivingReportInventory `json:"receiving_report_inventory"`
	//child3 ReceivingReportAttachment models.ReceivingReportAttachment `json:"receiving_report_attachment
}

func GetReceivingReports(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ReceivingReport          []models.ReceivingReport          `json:"receiving_report"`
		ReceivingReportDetails   []models.ReceivingReportDetails   `json:"receiving_report_details"` //child1
		ReceivingReportInventory []models.ReceivingReportInventory `json:"receiving_report_inventory"`
		//child3 ReceivingReportAttachment []models.ReceivingReportAttachment `json:"receiving_report_attachment"`
	}

	var response Response

	if err := services.DbGet(&response.ReceivingReport, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	if err := services.DbGet(&response.ReceivingReportDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report details")
	}

	if err := services.DbGet(&response.ReceivingReportInventory, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report inventory")
	}

	// if err := services.DbGet(&response.ReceivingReportAttachment, conditions); err != nil {
	// 	return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report attachment")
	// }

	return response, 0, nil
}

func GetReceivingReport(id int) (ReceivingReportBody, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var body ReceivingReportBody

	if err := services.DbGet(&body.ReceivingReport, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	//children condition
	conditions = map[string]interface{}{
		"receiving_report_id": body.ReceivingReport.ID, //id,
	}

	//panganay
	receivingReportDetails, _, err := GetReceivingReportDetails(conditions) //get list of children
	if err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	body.ReceivingReportDetails = receivingReportDetails

	//ampon
	receivingReportInventory, _, err := GetReceivingReportInventory(conditions)
	if err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	body.ReceivingReportInventory = receivingReportInventory

	//bunso
	// if err := services.GetInventory(&body.ReceivingReport, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	// }

	return body, 0, nil
}

func CreateReceivingReport(c *fiber.Ctx, tx *gorm.DB) (ReceivingReportBody, int, error) {
	var body ReceivingReportBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.ReceivingReport); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report")
	}

	generatedDocNo := utils.DocNoGenerator(
		body.ReceivingReport.ID)
	body.ReceivingReport.DOC = generatedDocNo

	if err := tx.Model(&body.ReceivingReport).Update("doc", body.ReceivingReport.DOC).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating receiving report doc")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	receivingreportat := models.ReceivingReportAt{
		RefId:                  body.ReceivingReport.ID,
		Code:                   body.ReceivingReport.DOC, //doc is literally only unique in rr aside from id ofc
		ReceivingReportContent: body.ReceivingReport.ReceivingReportContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &receivingreportat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	//child inventory (not needed in DGs since looping is needed)
	//body.ReceivingReportDetails.ReceivingReportId = body.ReceivingReport.ID //setting parent id as ref id (done in for loop)
	// body.WarehouseArea.WarehouseNameId = body.WarehouseName.ID //child2

	//child1
	for _, detail := range body.ReceivingReportDetails {
		detail.ReceivingReportId = body.ReceivingReport.ID
		if err := CreateReceivingReportDetails(tx, body.ReceivingReport.ID, detail, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//child 2
	for _, inventory := range body.ReceivingReportInventory {
		inventory.ReceivingReportId = body.ReceivingReport.ID
		if err := CreateReceivingReportInventory(tx, body.ReceivingReport.ID, inventory, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//child 3
	// if err := CreateReceivingReportFiles(tx, body.parent.ID, body.child, at); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

func UpdateReceivingReport(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (ReceivingReportBody, int, error) {
	var body ReceivingReportBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.ReceivingReport, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating receiving report")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata :=
		models.ReceivingReportAt{
			RefId:                  body.ReceivingReport.ID,
			ReceivingReportContent: body.ReceivingReport.ReceivingReportContent,
			At:                     at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	//childs
	// conditions = map[string]interface{}{
	// 	"warehouse_name_id": body.WarehouseName.ID,
	// }

	//panganay
	for i := range body.ReceivingReportDetails {
		body.ReceivingReportDetails[i].ReceivingReportId = body.ReceivingReport.ID

		if body.ReceivingReportDetails[i].ID == 0 {
			// Create if no id was found
			if err := CreateReceivingReportDetails(tx, body.ReceivingReport.ID, body.ReceivingReportDetails[i], at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		} else {
			// Update if id has nonzero value
			if err := UpdateReceivingReportDetails(tx, body.ReceivingReportDetails[i], at, conditions); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	}

	for i := range body.ReceivingReportInventory {
		body.ReceivingReportInventory[i].ReceivingReportId = body.ReceivingReport.ID

		if body.ReceivingReportInventory[i].ID == 0 {
			if err := CreateReceivingReportInventory(tx, body.ReceivingReport.ID, body.ReceivingReportInventory[i], at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		} else {
			if err := UpdateReceivingReportInventory(tx, body.ReceivingReportInventory[i], at, conditions); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	}

	//bunso
	// if err := UpdateWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

func DeleteReceivingReport(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (DeleteReceivingReportBody, int, error) {
	var body DeleteReceivingReportBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.ReceivingReport, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting receiving report")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata :=
		models.ReceivingReportAt{
			RefId:                  body.ReceivingReport.ID,
			ReceivingReportContent: body.ReceivingReport.ReceivingReportContent,
			At:                     at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	//kids
	// conditions = map[string]interface{}{
	// 	"warehouse_name_id": body.WarehouseName.ID,
	// }

	// for _, detail := range body.ReceivingReportDetails {
	// 	detail.ReceivingReportId = body.ReceivingReport.ID
	// 	if err := CreateReceivingReportDetails(tx, body.ReceivingReport.ID, detail, at); err != nil {
	// 		return body, fiber.StatusInternalServerError, err
	// 	}
	// }

	conditions = map[string]interface{}{
		"receiving_report_id": body.ReceivingReport.ID, //id,
	}

	if err := DeleteReceivingReportDetails(tx, body.ReceivingReportDetails, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeleteReceivingReportInventory(tx, body.ReceivingReportInventory, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// if err := DeleteWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

// obviously deletion for deets specific row
func DeleteReceivingReportDetailsRow(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (DeleteReceivingReportBody, int, error) {
	var body DeleteReceivingReportBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.ReceivingReportDetails, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting receiving report detail specific row")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata :=
		models.ReceivingReportDetailsAt{
			RefId:                         body.ReceivingReportDetails.ID,
			ReceivingReportDetailsContent: body.ReceivingReportDetails.ReceivingReportDetailsContent,
			At:                            at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report details	 at")
	}

	return body, fiber.StatusOK, nil
}

// for long ah name child 2
func DeleteReceivingReportInventoryRow(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (DeleteReceivingReportBody, int, error) {
	var body DeleteReceivingReportBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.ReceivingReportInventory, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting receiving report inventory specific row")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata :=
		models.ReceivingReportInventoryAt{
			RefId:                           body.ReceivingReportInventory.ID,
			ReceivingReportInventoryContent: body.ReceivingReportInventory.ReceivingReportInventoryContent,
			At:                              at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report inventory at")
	}

	return body, fiber.StatusOK, nil
}
