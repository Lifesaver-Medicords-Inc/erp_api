package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type DeleteReceivingReportBody2 struct {
	ReceivingReport        models.ReceivingReport2        `json:"receiving_report"`
	ReceivingReportDetails models.ReceivingReportDetails2 `json:"receiving_report_details"` //child1
	//child3 ReceivingReportAttachment models.ReceivingReportAttachment `json:"receiving_report_attachment
}

type ReceivingReportBody2 struct {
	ReceivingReport        models.ReceivingReport2          `json:"receiving_report"`
	ReceivingReportDetails []models.ReceivingReportDetails2 `json:"receiving_report_details"` //child1
	//child3 ReceivingReportAttachment models.ReceivingReportAttachment `json:"receiving_report_attachment
}

func GetReceivingReports2(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ReceivingReport        []models.ReceivingReport2        `json:"receiving_report"`
		ReceivingReportDetails []models.ReceivingReportDetails2 `json:"receiving_report_details"` //child1
		//child3 ReceivingReportAttachment []models.ReceivingReportAttachment `json:"receiving_report_attachment"`
	}

	var response Response

	if err := services.DbGet(&response.ReceivingReport, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	if err := services.DbGet(&response.ReceivingReportDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report details")
	}

	// if err := services.DbGet(&response.ReceivingReportAttachment, conditions); err != nil {
	// 	return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report attachment")
	// }

	return response, 0, nil
}

func GetReceivingReport2(id int) (ReceivingReportBody2, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var body ReceivingReportBody2

	if err := services.DbGet(&body.ReceivingReport, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	//children condition
	conditions = map[string]interface{}{
		"receiving_report_id": body.ReceivingReport.ID, //id,
	}

	//panganay
	receivingReportDetails, _, err := GetReceivingReportDetails2(conditions) //get list of children
	if err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	body.ReceivingReportDetails = receivingReportDetails

	//bunso
	// if err := services.GetInventory(&body.ReceivingReport, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	// }

	return body, 0, nil
}

func CreateReceivingReport2(c *fiber.Ctx, tx *gorm.DB) (ReceivingReportBody2, int, error) {

	var body ReceivingReportBody2

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.ReceivingReport); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report")
	}

	generatedDocNo := utils.DocNoGenerator(body.ReceivingReport.ID)
	body.ReceivingReport.DOC = generatedDocNo

	if err := tx.Model(&body.ReceivingReport).Update("doc", body.ReceivingReport.DOC).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating receiving report doc")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	receivingreportat := models.ReceivingReportAt2{
		RefId:                   body.ReceivingReport.ID,
		Code:                    body.ReceivingReport.DOC, //doc is literally only unique in rr aside from id ofc
		ReceivingReportContent2: body.ReceivingReport.ReceivingReportContent2,
		At:                      at,
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
		if err := CreateReceivingReportDetails2(tx, body.ReceivingReport.ID, detail, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//child 3
	// if err := CreateReceivingReportFiles(tx, body.parent.ID, body.child, at); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

func UpdateReceivingReport2(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (ReceivingReportBody2, int, error) {
	var body ReceivingReportBody2

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
		models.ReceivingReportAt2{
			RefId:                   body.ReceivingReport.ID,
			ReceivingReportContent2: body.ReceivingReport.ReceivingReportContent2,
			At:                      at,
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
			if err := CreateReceivingReportDetails2(tx, body.ReceivingReport.ID, body.ReceivingReportDetails[i], at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		} else {
			// Update if id has nonzero value
			if err := UpdateReceivingReportDetails2(tx, body.ReceivingReportDetails[i], at, conditions); err != nil {
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

func DeleteReceivingReport2(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (DeleteReceivingReportBody2, int, error) {
	var body DeleteReceivingReportBody2
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
		models.ReceivingReportAt2{
			RefId:                   body.ReceivingReport.ID,
			ReceivingReportContent2: body.ReceivingReport.ReceivingReportContent2,
			At:                      at,
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

	if err := DeleteReceivingReportDetails2(tx, body.ReceivingReportDetails, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// if err := DeleteWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

// obviously deletion for deets specific row
func DeleteReceivingReportDetailsRow2(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (DeleteReceivingReportBody2, int, error) {
	var body DeleteReceivingReportBody2
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
		models.ReceivingReportDetailsAt2{
			RefId:                          body.ReceivingReportDetails.ID,
			ReceivingReportDetailsContent2: body.ReceivingReportDetails.ReceivingReportDetailsContent2,
			At:                             at,
		}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report details	 at")
	}

	return body, fiber.StatusOK, nil
}
