package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

func GetInvTracker(c *fiber.Ctx) error {
	data, status, err := setup_services.GetInvTracker(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetInvName(c *fiber.Ctx) error {
	data, status, err := setup_services.GetInvName(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
