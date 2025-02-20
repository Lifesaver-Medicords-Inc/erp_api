package setup_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// GET PROJECT
func GetProjects(conditions map[string]interface{}) ([]models.Project, int, error) {
	var projects []models.Project

	if err := services.DbGet(&projects, conditions); err != nil {
		return projects, fiber.StatusInternalServerError, errors.New("failed getting projects")
	}

	return projects, 0, nil
}

// CREATE PROJECT
func CreateProject(c *fiber.Ctx, tx *gorm.DB) (models.Project, int, error) {
	var body models.Project
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating application")
		}

		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func UpdateProject(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Project, int, error) {
	var body models.Project
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("Error Parsingggggg", err)

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating project")
	}

	return body, 0, nil
}
