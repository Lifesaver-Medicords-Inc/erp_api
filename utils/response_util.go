package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// HasPrev and Total are additive - every existing caller sets HasNext/PageSize by field name
// and simply leaves these zero, and the clients deserialize by key. They exist for a screen
// that pages in BOTH directions (Item Entry's << PREV / NEXT >>), where "is there a page
// behind this one" cannot be inferred from a forward-only cursor.
type PaginationMeta struct {
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
	PageSize   int   `json:"page_size"`
	Page       int   `json:"page"`
	TotalPages int   `json:"total_pages"`
	Total      int64 `json:"total"`
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
