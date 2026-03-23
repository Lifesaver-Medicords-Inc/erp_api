package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type PaginationMeta struct {
	HasNext  bool `json:"has_next"`
	PageSize int  `json:"page_size"`
}

func RespondError(c *fiber.Ctx, status int, message string) error {

	log.Error("Exception Message", message)

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

func RespondSuccess(c *fiber.Ctx, data interface{}, pagination ...PaginationMeta) error {

	response := fiber.Map{
		"success": true,
		"data":    data,
	}

	if len(pagination) > 0 {
		response["pagination"] = pagination[0]
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
