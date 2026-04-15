package job_order_services

import (
	// "errors"

	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type JobOrderService struct{}

func NewJobOrderService() *JobOrderService {
	return &JobOrderService{}
}

func (s *JobOrderService) GetJobOrder(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.JobOrder

	if err := services.DbRaw(&response, "sp_GetJobOrders", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting job order data")
	}

	return response, 0, nil
}

func (s *JobOrderService) GetSalesOrderViewEng(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.SalesOrderViewBody

	if err := services.DbGet(&response.SalesOrderView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order")
	}

	if err := services.DbGet(&response.SalesOrderDetailsView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order details")
	}

	return response, 0, nil
}

func (s *JobOrderService) GetEngineerList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.EngineerListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting engineer list")
	}

	return response, 0, nil
}

func (s *JobOrderService) GetComponents(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.Components

	if err := services.DbRaw(&response, "sp_GetComponents", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item components data")
	}

	return response, 0, nil
}

func (s *JobOrderService) CreateJobOrder(body *[]models.JobOrder, at models.At) (*[]models.JobOrder, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	for i := range *body {
		item := &(*body)[i]

		if err := services.DbInsert(tx, item); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				return body, fiber.StatusInternalServerError, errors.New("duplicate record error")
			}
			return body, fiber.StatusInternalServerError, errors.New("failed creating job order")
		}

		atdata := models.JobOrderAt{
			RefId:           item.ID,
			JobOrderContent: item.JobOrderContent,
			At:              at,
		}

		if err := services.DbInsert(tx, &atdata); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, 0, nil
}

func (s *JobOrderService) UpdateJobOrder(body *[]models.JobOrder, conditions map[string]interface{}, at models.At) (*[]models.JobOrder, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	defer tx.Rollback()

	for i := range *body {
		item := &(*body)[i]

		itemConditions := map[string]interface{}{"id": item.ID}

		// Get the existing job order record
		var existing models.JobOrder
		if err := tx.Where(itemConditions).First(&existing).Error; err != nil {
			return body, fiber.StatusNotFound, fmt.Errorf("job order not found for id %v", item.ID)
		}

		// Handle report_base (replace file if new one is uploaded)
		if item.ReportBase != "" {
			if existing.ReportBase != "" {
				oldPath := "./files/" + existing.ReportBase
				if err := services.DeleteFile(oldPath); err != nil && !os.IsNotExist(err) {
					fmt.Println("Failed deleting old report file:", err)
				}
			}

			filename, err := services.UploadFile(item.ReportBase)
			if err != nil {
				return body, fiber.StatusInternalServerError, fmt.Errorf("failed saving report file for id %v", item.ID)
			}

			item.ReportBase = filename
		} else {
			item.ReportBase = existing.ReportBase
		}

		// Update record in DB
		if err := services.DbUpdate(tx, item, itemConditions); err != nil {
			return body, fiber.StatusInternalServerError, fmt.Errorf("failed updating job order for id %v", item.ID)
		}

		// Insert audit record
		atdata := models.JobOrderAt{
			RefId:           item.ID,
			JobOrderContent: item.JobOrderContent,
			At:              at,
		}

		if err := services.DbInsert(tx, &atdata); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, 0, nil
}
