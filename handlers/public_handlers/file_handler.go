package public_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

func DeleteFile(c *fiber.Ctx) error {
	type Request struct {
		Path string `json:"path"`
	}

	var request Request
	if err := c.BodyParser(&request); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	if err := services.DeleteFile(request.Path); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(c, "file deleted")
}
