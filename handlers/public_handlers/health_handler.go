package public_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/utils"
)

func CheckHealth(c *fiber.Ctx) error {
	return utils.RespondSuccess(c, nil)
}
