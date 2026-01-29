package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

type TaxSetupHandler struct {
	Service *setup_services.TaxSetupService
}

func NewTaxSetupHandler(service *setup_services.TaxSetupService) *TaxSetupHandler {
	return &TaxSetupHandler{Service: service}
}

func (h *TaxSetupHandler) GetTaxSetup(c *fiber.Ctx) error {

	data, status, err := h.Service.GetTaxSetup(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *TaxSetupHandler) GetChartOfAccountSetup(c *fiber.Ctx) error {

	data, status, err := h.Service.GetChartOfAccountSetup(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *TaxSetupHandler) GetTaxClassificationSetup(c *fiber.Ctx) error {

	codeParams := c.Params("code")

	data, status, err := h.Service.GetTaxClassificationSetup(codeParams)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *TaxSetupHandler) CreateTaxSetup(c *fiber.Ctx) error {
	var body accounting_models.TaxSetupBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateTaxSetup(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *TaxSetupHandler) UpdateTaxSetup(c *fiber.Ctx) error {
	var body accounting_models.TaxSetupBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateTaxSetup(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *TaxSetupHandler) DeleteTaxSetup(c *fiber.Ctx) error {
	var body accounting_models.TaxSetupBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteTaxSetup(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
