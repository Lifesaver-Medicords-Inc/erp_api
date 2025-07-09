package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

func GetTaxSetup(c *fiber.Ctx) error {

	data, status, err := setup_services.GetTaxSetup(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateTaxSetup(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transactions")
	}
	data, status, err := setup_services.CreateTaxSetup(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transactions")
	}
	return utils.RespondSuccess(c, data)
}
