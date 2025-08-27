package adminhandlers

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

func GetAllUsers(c *fiber.Ctx) error {
	data, status, err := adminservices.GetUsers(nil)

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

	data, status, err := adminservices.CreateUser(c, tx)
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

type UpdatePositionRequest struct {
	Id         int `json:"id"`
	PositionId int `json:"position_id"`
}

func UpdateUserPosition(c *fiber.Ctx) error {
	// Parse JSON body
	var req UpdatePositionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Begin transaction
	tx := initializers.DB.Begin()

	conditions := map[string]interface{}{
		"id": req.Id,
	}

	user, _, err := adminservices.GetUser(conditions)

	if err != nil {
		return utils.RespondError(c, fiber.StatusNotFound, "User not found")
	}

	user.PositionId = uint(req.PositionId)

	// Marshal the updated user into JSON
	jsonBody, err := json.Marshal(user)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to marshal user")
	}

	// Replace the request body with the marshaled JSON
	c.Request().SetBody(jsonBody)

	data, status, err := adminservices.UpdateUser(c, tx, conditions)

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
