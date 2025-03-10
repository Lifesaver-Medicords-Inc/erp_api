package purchasing_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetPurchasingList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.PurchasingListView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchasing list")
	}
	return response, 0, nil
}
