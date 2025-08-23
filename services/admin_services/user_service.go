package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetUsers(conditions map[string]interface{}) ([]models.User, int, error) {
	var users []models.User

	if err := services.DbGet(&users, conditions); err != nil {
		return users, fiber.StatusInternalServerError, errors.New("failed getting users")
	}

	for i := range users {
		users[i].Password = ""
	}

	return users, 0, nil
}
