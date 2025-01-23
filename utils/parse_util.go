package utils

import "github.com/gofiber/fiber/v2"

func ParseBody(c *fiber.Ctx, body interface{}) error {
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, "Cannot bind request")
	}

	return nil
}
