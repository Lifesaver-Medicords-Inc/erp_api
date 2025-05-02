package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func RespondError(c *fiber.Ctx, status int, message string) error {

	log.Error("Exception Message", message)

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

func RespondSuccess(c *fiber.Ctx, data interface{}) error {

	log.Infof("SUCCESS: %s", data)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}
