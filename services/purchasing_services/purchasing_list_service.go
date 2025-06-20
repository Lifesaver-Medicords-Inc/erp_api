package purchasing_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetSOPurchasingList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.SOPurchasingListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasing list")
	}
	return response, 0, nil
}
func GetPRPurchasingList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.PRPurchasingListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasing list")
	}
	return response, 0, nil
}
func GetSOPurchasingListSupplier(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.PurchasingListSupplierView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasing list supplier")
	}
	return response, 0, nil
}

func GetPurchasingGuidingPrice(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.PurchasingGuidingPriceView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasing guiding price")
	}
	return response, 0, nil
}
