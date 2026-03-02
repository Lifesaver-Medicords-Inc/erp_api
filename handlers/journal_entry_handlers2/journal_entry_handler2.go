package journal_entry_handlers2

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
	"github.com/pierceperado/smpc/utils"
)

type JournalEntryHandler2 struct {
	Service *journal_entry_services2.JournalEntryService2
}

func NewJournalEntryHandler2(service *journal_entry_services2.JournalEntryService2) *JournalEntryHandler2 {
	return &JournalEntryHandler2{Service: service}
}

func (h *JournalEntryHandler2) GetCompanySetup(c *fiber.Ctx) error {

	conditions := map[string]interface{}{
		"id": 1,
	}

	data, status, err := h.Service.GetCompanySetup(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JournalEntryHandler2) GetCurrentJournal(c *fiber.Ctx) error {
	data, status, err := h.Service.GetCurrentJournal(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JournalEntryHandler2) GetJournalEntry(c *fiber.Ctx) error {
	data, status, err := h.Service.GetJournalEntry(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JournalEntryHandler2) CreateJournalEntry(c *fiber.Ctx) error {
	var body accounting_models.JournalEntryBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateJournalEntry(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JournalEntryHandler2) UpdateJournalEntry(c *fiber.Ctx) error {
	var body accounting_models.JournalEntryBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateJournalEntry(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JournalEntryHandler2) DeleteJournalEntry(c *fiber.Ctx) error {
	var body accounting_models.JournalEntryBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteJournalEntry(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
