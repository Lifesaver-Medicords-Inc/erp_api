package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetInvTracker(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.InvTrackerView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item inventory list")
	}

	return response, 0, nil
}

func GetInvName(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.InvNameView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item inventory name")
	}

	return response, 0, nil
}
