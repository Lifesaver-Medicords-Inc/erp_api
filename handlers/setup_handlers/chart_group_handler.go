package setup_handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

// TWO HANDLERS FOR GETTING ChartGroupS
func GetChartGroups(c *fiber.Ctx) error {
	data, status, err := setup_services.GetChartGroups(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetChartGroup(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	fmt.Println("Notes:GetChartGroup")
	data, status, err := setup_services.GetChartGroup(idNum)
	fmt.Println("Notes:GetChartGroup 1")
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

///////////////////////////////

// HANDLER FOR CREATING ChartGroup
func CreateChartGroup(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := setup_services.CreateChartGroup(c, tx)
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

// HANDLER FOR UPDATING ChartGroup
func UpdateChartGroup(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.UpdateChartGroup(c, tx, nil)
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

// HANDLER DELETE ChartGroup
func DeleteChartGroup(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.DeleteChartGroup(c, tx, nil)
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
