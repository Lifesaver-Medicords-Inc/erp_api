package engineering_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetEngineeringRedboxQuotationList(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		EngineeringQuotationList []models.EngineeringRedboxQuotationListView `json:"quotationlist"`
	}

	var response Response

	if err := services.DbGet(&response.EngineeringQuotationList, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get quotation list")
	}

	return response, 0, nil
}

func GetSortedEngineeringRedboxQuotationList(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		EngineeringQuotationList []models.EngineeringRedboxQuotationListView `json:"quotationlist"`
	}

	var response Response

	// Fetch data without sorting
	if err := services.DbGet(&response.EngineeringQuotationList, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get quotation list")
	}

	return response, 0, nil
}

func GetEngineeringRedboxJobOrder(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		EngineeringJobOrder []models.EngineeringRedboxJobOrderView `json:"joborder"`
	}

	var response Response

	if err := services.DbGet(&response.EngineeringJobOrder, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get job orders")
	}

	return response, 0, nil
}

func GetSortedEngineeringRedboxJobOrder(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		EngineeringJobOrder []models.EngineeringRedboxJobOrderView `json:"joborder"`
	}

	var response Response

	// Call stored procedure with caching
	if err := services.DbRaw(&response.EngineeringJobOrder, "sp_GetEngineeringRedboxJobOrder", conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get job order")
	}

	return response, fiber.StatusOK, nil
}
