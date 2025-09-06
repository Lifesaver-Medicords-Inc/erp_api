package adminhandlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/services/public_services"
	"github.com/pierceperado/smpc/utils"
)

func GetAllUsers(c *fiber.Ctx) error {
	id := c.Query("id")
	firstName := c.Query("first-name")
	lastName := c.Query("last-name")
	department := c.Query("department")

	conditions := make(map[string]interface{})

	idNum, _ := strconv.Atoi(id)

	if idNum != 0 {
		conditions["id"] = id
	}

	if firstName != "" {
		conditions["first-name"] = firstName
	}

	if lastName != "" {
		conditions["last-name"] = lastName
	}

	if department != "" {
		conditions["department"] = department
	}

	data, status, err := adminservices.GetUsers(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetUser(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	data, status, err := adminservices.GetUsers(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetPositionUsers(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"position_id": idNum,
	}

	data, status, err := adminservices.GetUsers(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateUser(c *fiber.Ctx) error {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := public_services.CreateAccount(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdateUser(c *fiber.Ctx) error {
	// Parse JSON body
	var body models.User
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	conditions := map[string]interface{}{
		"id": body.ID,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := adminservices.UpdateUser(body, conditions, at)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func UpdateUserPosition(c *fiber.Ctx) error {
	// Parse JSON body
	var body models.User
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	conditions := map[string]interface{}{
		"id": body.ID,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := adminservices.UpdateUser(body, conditions, at)

	if err != nil {

		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	conditions := map[string]interface{}{"id": idNum}

	user, status, err := adminservices.GetUser(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	permissionConditions := map[string]interface{}{
		"user_id": user.ID,
	}

	adminservices.DeletePermission(permissionConditions, at)

	data, status, err := adminservices.DeleteUser(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
