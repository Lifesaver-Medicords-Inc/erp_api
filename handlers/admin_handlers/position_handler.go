package adminhandlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	adminmodels "github.com/pierceperado/smpc/models/admin_models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type PositionWithAccess struct {
	models.Position
	Access []*adminmodels.PositionAccess `json:"access"`
}

func GetPositions(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	positions, status, err := adminservices.GetPositions(nil, tx)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	var response []PositionWithAccess

	for _, pos := range positions {
		accessCondition := map[string]interface{}{
			"position_id": pos.ID,
		}

		access, _, err := adminservices.GetPositionAccess(accessCondition)
		if err != nil {
			access = []adminmodels.PositionAccess{}
		}

		// Append PositionWithAccess to response
		response = append(response, PositionWithAccess{
			Position: pos,
			Access:   adminservices.ToPtrSlice(access),
		})
	}

	return utils.RespondSuccess(c, response)
}

func GetPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	tx := initializers.DB.Begin()
	position, status, err := adminservices.GetPosition(conditions, tx)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	// Get access for this position
	accessCondition := map[string]interface{}{
		"position_id": position.ID,
	}
	access, _, err := adminservices.GetPositionAccess(accessCondition)
	if err != nil {
		access = []adminmodels.PositionAccess{}
	}

	// Wrap into response struct with access field
	response := PositionWithAccess{
		Position: position,
		Access:   adminservices.ToPtrSlice(access),
	}

	return utils.RespondSuccess(c, response)
}

func CreatePosition(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := adminservices.CreatePosition(c, tx)
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

func UpdatePosition(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := adminservices.UpdatePosition(c, tx, nil)
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

func DeletePosition(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := adminservices.DeletePosition(c, tx, nil)
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
