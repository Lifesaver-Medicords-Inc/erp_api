package adminhandlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
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

func GetPositionUsers(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}
	tx := initializers.DB.Begin()
	// conditions := map[string]interface{}{
	// 	"position_id": idNum,
	// }

	// data, status, err := adminservices.GetUsers(conditions)

	// if err != nil {
	// 	return utils.RespondError(c, status, err.Error())
	// }

	var positions []models.User

	query := tx.Model(&models.User{}).Where("position_id = ?", idNum)

	if err := query.Find(&positions).Error; err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Failed getting users")
	}

	return utils.RespondSuccess(c, positions)
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

	fmt.Print("%v - %v", req.Id, req.PositionId)
	// Perform the update
	result := tx.Model(&models.User{}).
		Where("id = ?", req.Id).
		Update("position_id", req.PositionId)

	// Handle error
	if result.Error != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed updating user position")
	}

	// Optional: check if any row was actually updated
	if result.RowsAffected == 0 {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusNotFound, "User not found")
	}

	// Commit and return success
	tx.Commit()
	return utils.RespondSuccess(c, "User position updated successfully")
}
