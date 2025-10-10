package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetInvLogbook(conditions map[string]interface{}) (interface{}, int, error) {

	var response []models.InvLogbookView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item inventory logbook")
	}

	//Invalidate cache
	InvalidateItemCaches()

	return response, 0, nil
}
