package job_order_services

import (
	// "errors"

	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetJobOrder(userId int64) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"UserId": userId,
	}
	var response []models.JobOrderView

	if err := services.DbRaw(&response, "sp_GetJobOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting job order data")
	}

	return response, 0, nil
}

func GetSalesJobOrder(salesOrder string) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"SalesOrder": salesOrder,
	}
	var response []models.JobOrderSales

	if err := services.DbRaw(&response, "sp_GetSalesOrdersFromJobOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting job order sales list data")
	}

	return response, 0, nil
}

func GetAllSalesJobOrder(userId int64) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"UserId": userId,
	}
	var response []models.JobOrderSales

	if err := services.DbRaw(&response, "sp_GetAllSalesOrdersFromJobOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting job order sales list data")
	}

	return response, 0, nil
}

func GetSalesDetailsJobOrder(orderId int64) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"OrderId": orderId,
	}
	var response []models.JobOrderSalesDetails

	if err := services.DbRaw(&response, "sp_GetSalesOrderDetailsFromJobOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting job order sales details list data")
	}

	return response, 0, nil
}

func GetComponents(bomId int64) (interface{}, int, error) {
	conditions := map[string]interface{}{
		"BomId": bomId,
	}
	var response []models.Components

	if err := services.DbRaw(&response, "sp_GetComponents", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item components data")
	}

	return response, 0, nil
}

func CreateJobOrder(c *fiber.Ctx, tx *gorm.DB) (models.JobOrder, int, error) {
	var body models.JobOrder

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating job order list")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	fmt.Println("at  ok ", at)

	if !ok {
		at = models.At{}
		fmt.Println("at not ok ", at)

	}

	atdata := models.JobOrderAt{RefId: body.JobOrderID, JobOrderContent: models.JobOrderContent{ItemDesc: body.ItemDesc, Type: body.Type}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating jobOrderListAt")
	}

	if err := services.InvalidateCacheByModel(models.JobOrderView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, 0, nil
}

func UpdateJobOrder(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.JobOrder, int, error) {
	var body models.JobOrder

	if err := c.BodyParser(&body); err != nil {
		fmt.Println("This is the error:", err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// 🔹 Step 1: Get the existing job order record
	var existing models.JobOrder
	if err := tx.Where(conditions).First(&existing).Error; err != nil {
		return body, fiber.StatusNotFound, errors.New("job order not found")
	}

	// 🔹 Step 2: Handle report_base (replace file if new one is uploaded)
	if body.ReportBase != "" {
		// Delete old file if it exists
		if existing.ReportBase != "" {
			oldPath := fmt.Sprintf("./files/%s", existing.ReportBase)
			if err := services.DeleteFile(oldPath); err != nil && !os.IsNotExist(err) {
				fmt.Println("Failed deleting old report file:", err)
			}
		}

		// Save new file
		filename, err := services.UploadFile(body.ReportBase)
		if err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed saving report file")
		}

		// Store new filename in DB instead of raw base64
		body.ReportBase = filename
	} else {
		// Keep the old filename if no new file is uploaded
		body.ReportBase = existing.ReportBase
	}

	// 🔹 Step 3: Update record in DB
	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating Job Order List")
	}

	// 🔹 Step 4: Save audit trail
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.JobOrderAt{
		RefId: body.JobOrderID,
		JobOrderContent: models.JobOrderContent{
			Type:     body.Type,
			ItemDesc: body.ItemDesc,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating Job Order Listat")
	}

	// 🔹 Step 5: Invalidate cache
	if err := services.InvalidateCacheByModel(models.JobOrderView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	fmt.Println("JOB ORDER UPDATE: ", body)

	return body, 0, nil
}
