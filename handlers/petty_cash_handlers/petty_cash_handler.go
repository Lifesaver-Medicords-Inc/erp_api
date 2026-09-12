package petty_cash_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/petty_cash_services"
	"github.com/pierceperado/smpc/utils"
)

// Petty Cash Replenishment (spec 5.26).
type PettyCashHandler struct {
	Service *petty_cash_services.PettyCashService
}

func NewPettyCashHandler(service *petty_cash_services.PettyCashService) *PettyCashHandler {
	return &PettyCashHandler{Service: service}
}

func at(c *fiber.Ctx) models.At {
	value, ok := c.Locals("at").(models.At)
	if !ok {
		return models.At{}
	}
	return value
}

func (h *PettyCashHandler) GetAll(c *fiber.Ctx) error {
	rows, status, err := h.Service.GetAll()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, rows)
}

func (h *PettyCashHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	row, status, serviceErr := h.Service.Get(uint(id))
	if serviceErr != nil {
		return utils.RespondError(c, status, serviceErr.Error())
	}
	return utils.RespondSuccess(c, row)
}

// The next cycle, as a copy of the last with its rows cleared and the balance
// carried forward. Nothing is stored until it is saved.
func (h *PettyCashHandler) NextCycle(c *fiber.Ctx) error {
	row, status, err := h.Service.NextCycle()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, row)
}

func (h *PettyCashHandler) Save(c *fiber.Ctx) error {
	var body accounting_models.PettyCash
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	status, err := h.Service.Save(&body, at(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, body)
}

type approveBody struct {
	ID           uint   `json:"id"`
	ApprovedBy   string `json:"approved_by"`
	ApprovedDate string `json:"approved_date"`
}

func (h *PettyCashHandler) Approve(c *fiber.Ctx) error {
	var body approveBody
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}
	if body.ID == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "id is required")
	}

	status, err := h.Service.Approve(body.ID, body.ApprovedBy, body.ApprovedDate, at(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, body)
}

func (h *PettyCashHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	status, serviceErr := h.Service.Delete(uint(id), at(c))
	if serviceErr != nil {
		return utils.RespondError(c, status, serviceErr.Error())
	}
	return utils.RespondSuccess(c, id)
}
