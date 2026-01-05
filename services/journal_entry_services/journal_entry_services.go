package journal_entry_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type Body struct {
	accounting_models.JournalEntry
	JournalEntryDetails []accounting_models.JournalEntryDetails `json:"journal_entry_details"`
}

type JournalEntryBody struct {
	JournalEntry        accounting_models.JournalEntry          `json:"journal_entry"`
	JournalEntryDetails []accounting_models.JournalEntryDetails `json:"journal_entry_details"`
}

type JournalEntryGet struct {
	JournalEntry        []accounting_models.JournalEntry        `json:"journal_entry"`
	JournalEntryDetails []accounting_models.JournalEntryDetails `json:"journal_entry_details"`
}

func GetJournalEntries(conditions map[string]interface{}) (interface{}, int, error) {
	var response JournalEntryGet

	if err := services.DbGet(&response.JournalEntry, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting journal entry")
	}

	if err := services.DbGet(&response.JournalEntryDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting journal entry details")
	}

	return response, 0, nil
}

func CreateJournals(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.JournalEntry); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to insert journal header")
	}

	at, ok := c.Locals("at").(models.At)
	userAt := utils.GetAtData(c, models.At{})
	at.AtUserId = userAt.AtUserId
	if !ok {
		at = models.At{}
	}

	parentChildAt := accounting_models.JournalEntryAt{RefId: body.ID, JournalEntryContent: body.JournalEntryContent, At: at}
	if err := services.DbInsert(tx, &parentChildAt); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating  journal header parentat")
	}

	return body, 0, nil
}
