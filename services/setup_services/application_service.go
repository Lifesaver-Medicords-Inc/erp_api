package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// GET APPLICATIONS
func GetApplications(conditions map[string]interface{}) ([]models.Application, int, error) {
	var application []models.Application

	if err := services.DbGet(&application, conditions); err != nil {
		return application, fiber.StatusInternalServerError, errors.New("failed getting applications")
	}

	return application, 0, nil
}

// GET APPLICATION BY ID
func GetApplication(id int) (models.Application, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var application models.Application

	if err := services.DbGet(&application, conditions); err != nil {
		return application, fiber.StatusInternalServerError, errors.New("failed getting application")
	}

	return application, 0, nil
}

// CREATE APPLICATION
func CreateApplication(c *fiber.Ctx, tx *gorm.DB) (models.Application, int, error) {
	var body models.Application
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

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ApplicationAt{RefId: body.ID, Code: body.Code, ApplicationContent: body.ApplicationContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating ApplicationAt")
	}

	return body, 0, nil
}

// UPDATE APPLICATION
func UpdateApplication(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Application, int, error) {
	var body models.Application
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating application")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ApplicationAt{RefId: body.ID, Code: body.Code, ApplicationContent: body.ApplicationContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating applicationAt")
	}

	return body, 0, nil
}

// DELETE APPLICATION
func DeleteApplication(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Application, int, error) {
	var body models.Application
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting application")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ApplicationAt{RefId: body.ID, Code: body.Code, ApplicationContent: body.ApplicationContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating ApplicationAt")
	}

	return body, 0, nil
}
