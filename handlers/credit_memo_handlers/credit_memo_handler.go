package credit_memo_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/credit_memo_services"
	"github.com/pierceperado/smpc/utils"
)

type CreditMemoHandler struct {
	Service *credit_memo_services.CreditMemoService
}

func NewCreditMemoHandler(service *credit_memo_services.CreditMemoService) *CreditMemoHandler {
	return &CreditMemoHandler{Service: service}
}

func (h *CreditMemoHandler) GetCreditMemo(c *fiber.Ctx) error {
	data, status, err := h.Service.GetCreditMemo(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *CreditMemoHandler) GetCreditMemoById(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid id is required")
	}

	data, status, err := h.Service.GetCreditMemo(map[string]interface{}{"ID": idNum})
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *CreditMemoHandler) CreateCreditMemo(c *fiber.Ctx) error {
	var body models.CreditMemoBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateCreditMemo(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

// ApproveCreditMemo - COO only, customer-side Credit Memos only. See
// CreditMemoService.ApproveCreditMemo's doc comment.
func (h *CreditMemoHandler) ApproveCreditMemo(c *fiber.Ctx) error {
	creditMemoId, err := strconv.Atoi(c.Params("id"))
	if err != nil || creditMemoId <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid credit memo id is required")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	actingUserId := uint(0)
	if id, err := strconv.Atoi(at.AtUserId); err == nil && id > 0 {
		actingUserId = uint(id)
	}

	status, err := h.Service.ApproveCreditMemo(uint(creditMemoId), actingUserId, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, nil)
}
