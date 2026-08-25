package debit_memo_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/debit_memo_services"
	"github.com/pierceperado/smpc/utils"
)

type DebitMemoHandler struct {
	Service *debit_memo_services.DebitMemoService
}

func NewDebitMemoHandler(service *debit_memo_services.DebitMemoService) *DebitMemoHandler {
	return &DebitMemoHandler{Service: service}
}

func (h *DebitMemoHandler) GetDebitMemo(c *fiber.Ctx) error {
	data, status, err := h.Service.GetDebitMemo(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *DebitMemoHandler) GetDebitMemoById(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid id is required")
	}

	data, status, err := h.Service.GetDebitMemo(map[string]interface{}{"ID": idNum})
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *DebitMemoHandler) CreateDebitMemo(c *fiber.Ctx) error {
	var body models.DebitMemoBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateDebitMemo(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
