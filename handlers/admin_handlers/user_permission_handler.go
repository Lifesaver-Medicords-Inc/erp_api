package adminhandlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

func GetPermission(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}
	conditions := map[string]interface{}{
		"id": idNum,
	}

	data, status, err := adminservices.GetPermission(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetPermissions(c *fiber.Ctx) error {

	data, status, err := adminservices.GetPermissions(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetUserPermission(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}
	conditions := map[string]interface{}{
		"user_id": idNum,
	}

	data, status, err := adminservices.GetPermission(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreatePermission(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	var body models.UserPermission
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid input")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, _, err := adminservices.CreatePermission(body, at)

	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdatePermission(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	var body models.UserPermission
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid input")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := adminservices.UpdatePermission(body, conditions, at)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func DeletePermission(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := adminservices.DeletePermission(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
