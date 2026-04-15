package setup_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

// TWO HANDLERS FOR GETTING GetGeneralLedgerMappers
func GetGeneralLedgerMappers(c *fiber.Ctx) error {
	data, status, err := setup_services.GetGeneralLedgerMappers(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetGeneralLedgerMapper(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := setup_services.GetGeneralLedgerMapper(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// HANDLER FOR UPDATING GeneralLedgerMappers
func UpdateGeneralLedgerMappers(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	status, err := setup_services.UpdateGeneralLedgerMapper(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, true)
}
