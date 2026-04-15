package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

type ChartOfAccountHandler struct {
	Service *setup_services.ChartOfAccountService
}

func NewChartOfAccountHandler(service *setup_services.ChartOfAccountService) *ChartOfAccountHandler {
	return &ChartOfAccountHandler{Service: service}
}

func (h *ChartOfAccountHandler) GetChartOfAccounts(c *fiber.Ctx) error {
	data, status, err := h.Service.GetChartOfAccounts(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ChartOfAccountHandler) CreateChartOfAccount(c *fiber.Ctx) error {
	var body accounting_models.ChartOfAccounts

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateChartOfAccounts(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ChartOfAccountHandler) UpdateChartOfAccount(c *fiber.Ctx) error {
	var body accounting_models.ChartOfAccounts

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateChartOfAccount(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ChartOfAccountHandler) DeleteChartOfAccount(c *fiber.Ctx) error {
	var body accounting_models.ChartOfAccounts

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteChartOfAccount(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
