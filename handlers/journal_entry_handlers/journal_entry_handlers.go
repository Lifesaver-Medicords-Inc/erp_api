package journal_entry_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/journal_entry_services.go"
	"github.com/pierceperado/smpc/utils"
)

func CreateJournalEntries(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transactions")
	}

	data, status, err := journal_entry_services.CreateJournals(c, tx)
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
