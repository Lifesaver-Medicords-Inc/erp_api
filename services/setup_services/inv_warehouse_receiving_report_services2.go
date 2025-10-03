package setup_services

import (
	"errors"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type DeleteReceivingReportBody2 struct {
	ReceivingReport        models.ReceivingReport2        `json:"receiving_report"`
	ReceivingReportDetails models.ReceivingReportDetails2 `json:"receiving_report_details"` //child1
	ReceivingReportHistory models.ReceivingHistory        `json:"receiving_report_history"` //child1
}

type ReceivingReportBody2 struct {
	ReceivingReport        models.ReceivingReport2          `json:"receiving_report"`
	ReceivingReportDetails []models.ReceivingReportDetails2 `json:"receiving_report_details"` //child1
}

func GetReceivingReports2(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ReceivingReport        []models.ReceivingReport2        `json:"receiving_report"`
		ReceivingReportDetails []models.ReceivingReportDetails2 `json:"receiving_report_details"` //child1
		ReceivingReportHistory []models.ReceivingHistory        `json:"receiving_report_history"` //child2
	}

	var response Response

	if err := services.DbGet(&response.ReceivingReport, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	if err := services.DbGet(&response.ReceivingReportDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report details")
	}

	if err := services.DbGet(&response.ReceivingReportHistory, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report details")
	}

	InvalidateItemCaches()

	return response, 0, nil
}

func GetPurchaseOrderView(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.PurchaseOrderView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item purchase order")
	}

	// Sort by ID ascending
	sort.Slice(response, func(i, j int) bool {
		return response[i].Id < response[j].Id
	})

	//Invalidate cache
	InvalidateItemCaches()

	return response, 0, nil
}

func GetPurchaseOrderDetails(poId int64) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"PoId": poId,
	}
	var response []models.PurchaseOrderDetailsView

	if err := services.DbRaw(&response, "sp_GetPurchaseOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report details data")
	}

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

	receivingReportDetails, _, err := GetReceivingReportDetails2(conditions) //get list of children
	if err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	body.ReceivingReportDetails = receivingReportDetails

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
		Code:                    body.ReceivingReport.DOC,
		ReceivingReportContent2: body.ReceivingReport.ReceivingReportContent2,
		At:                      at,
	}

	if err := services.DbInsert(tx, &receivingreportat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	//child1
	for _, detail := range body.ReceivingReportDetails {
		detail.ReceivingReportId = body.ReceivingReport.ID
		if err := CreateReceivingReportDetails2(tx, body.ReceivingReport.ID, body.ReceivingReport.DateReceived, body.ReceivingReport.PurchaseOrderID, detail, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

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

	for i := range body.ReceivingReportDetails {
		body.ReceivingReportDetails[i].ReceivingReportId = body.ReceivingReport.ID

		if body.ReceivingReportDetails[i].ID == 0 {
			// Create if no id was found
			if err := CreateReceivingReportDetails2(tx, body.ReceivingReport.ID, body.ReceivingReport.DateReceived, body.ReceivingReport.PurchaseOrderID, body.ReceivingReportDetails[i], at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		} else {
			// Update if id has nonzero value
			if err := UpdateReceivingReportDetails2(tx, body.ReceivingReportDetails[i], at, conditions, body.ReceivingReport.ID, body.ReceivingReport.DateReceived, body.ReceivingReport.PurchaseOrderID); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	}

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

	conditions = map[string]interface{}{
		"receiving_report_id": body.ReceivingReport.ID, //id,
	}

	if err := DeleteReceivingReportHistory(tx, body.ReceivingReport.ID, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := DeleteReceivingReportDetails2(tx, body.ReceivingReportDetails, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

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
