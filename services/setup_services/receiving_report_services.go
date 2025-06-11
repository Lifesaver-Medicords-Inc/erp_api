package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type ReceivingReportBody struct {
	ReceivingReport models.ReceivingReport `json:"receiving_report"`
	//child1 ReceivingReportMain models.ReceivingReportMain `json:"receiving_report_main"`
	//child2 ReceivingReportInventory models.ReceivingReportInventory `json:"receiving_report_inventory"`
	//child3 ReceivingReportAttachment models.ReceivingReportAttachment `json:"receiving_report_attachment"`
}

type GetSpecificReivingReportBody struct {
	ReceivingReport models.ReceivingReport `json:"receiving_report"`
	//child1 ReceivingReportMain []models.ReceivingReportMain `json:"receiving_report_main"`
	//child2 ReceivingReportInventory []models.ReceivingReportInventory `json:"receiving_report_inventory"`
	//child3 ReceivingReportAttachment models.ReceivingReportAttachment `json:"receiving_report_attachment
}

func GetReceivingReports(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ReceivingReport []models.ReceivingReport `json:"receiving_report"`
		//child1 ReceivingReportMain []models.ReceivingReportMain `json:"receiving_report_main"`
		//child2 ReceivingReportInventory []models.ReceivingReportInventory `json:"receiving_report_inventory"`
		//child3 ReceivingReportAttachment []models.ReceivingReportAttachment `json:"receiving_report_attachment"`
	}

	var response Response

	if err := services.DbGet(&response.ReceivingReport, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	// if err := services.DbGet(&response.ReceivingReportMain, conditions); err != nil {
	// 	return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report main")
	// }

	// if err := services.DbGet(&response.ReceivingReportInventory, conditions); err != nil {
	// 	return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report inventory")
	// }

	// if err := services.DbGet(&response.ReceivingReportAttachment, conditions); err != nil {
	// 	return response, fiber.StatusInternalServerError, errors.New("failed getting receiving report attachment")
	// }

	return response, 0, nil
}

func GetReceivingReport(id int) (GetSpecificReivingReportBody, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var body GetSpecificReivingReportBody

	if err := services.DbGet(&body.ReceivingReport, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting receiving report")
	}

	//children
	// conditions = map[string]interface{}{
	// 	"receiving_report_id": body.ReceivingReport.ID,
	// }

	// if err := GetWarehouseAddress(&body.WarehouseAddress, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	//do this to Main (control panel child)
	// areas, _, err := GetWarehouseAreasDetached(conditions)
	// if err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }
	// body.WarehouseArea = areas

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
		Code:                   body.ReceivingReport.DOC, //doc is literally only unique in rr
		ReceivingReportContent: body.ReceivingReport.ReceivingReportContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &receivingreportat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating receiving report at")
	}

	//childern
	// body.WarehouseAddress.WarehouseNameId = body.WarehouseName.ID
	// body.WarehouseArea.WarehouseNameId = body.WarehouseName.ID

	// if err := CreateWarehouseAddress(tx, body.WarehouseName.ID, body.WarehouseAddress, at); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	// if err := CreateWarehouseArea(tx, body.WarehouseName.ID, body.WarehouseArea, at); err != nil {
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

	// if err := UpdateWarehouseAddress(tx, body.WarehouseAddress, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	// if err := UpdateWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}

func DeleteReceivingReport(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (ReceivingReportBody, int, error) {
	var body ReceivingReportBody
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

	// if err := DeleteWarehouseAddress(tx, body.WarehouseAddress, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	// if err := DeleteWarehouseArea(tx, body.WarehouseArea, at, conditions); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}
