package public_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

func ImageUpload(c *fiber.Ctx) error {
	type Request struct {
		File string `json:"file"`
	}

	var request Request
	if err := c.BodyParser(&request); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	path, err := services.UploadFile(request.File)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(c, path)
}
