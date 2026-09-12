package billing_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/billing_services"
	"github.com/pierceperado/smpc/utils"
)

// A/R Record of Transactions (spec 12.3), the Billing list (12.5) and the SO's
// billing ledger (12.4).
type BillingHandler struct {
	Service *billing_services.BillingService
}

func NewBillingHandler(service *billing_services.BillingService) *BillingHandler {
	return &BillingHandler{Service: service}
}

func at(c *fiber.Ctx) models.At {
	value, ok := c.Locals("at").(models.At)
	if !ok {
		return models.At{}
	}
	return value
}

func (h *BillingHandler) GetARRecords(c *fiber.Ctx) error {
	rows, status, err := h.Service.GetARRecords()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, rows)
}

type nextDueBody struct {
	SalesInvoiceId uint   `json:"sales_invoice_id"`
	NextDue        string `json:"next_due"`
}

// NEXT DUE is typed by A/R, never computed from the payment term (12.3).
func (h *BillingHandler) SetNextDue(c *fiber.Ctx) error {
	var body nextDueBody
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}
	if body.SalesInvoiceId == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "sales_invoice_id is required")
	}

	status, err := h.Service.SetNextDue(body.SalesInvoiceId, body.NextDue, at(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, body)
}

func (h *BillingHandler) GetBilling(c *fiber.Ctx) error {
	rows, status, err := h.Service.GetBilling()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, rows)
}

func (h *BillingHandler) GetTransactions(c *fiber.Ctx) error {
	rows, status, err := h.Service.GetTransactions(c.Query("so_no"))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, rows)
}

func (h *BillingHandler) CreateTransaction(c *fiber.Ctx) error {
	var body accounting_models.SoBillingTransaction
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	status, err := h.Service.CreateTransaction(&body, at(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, body)
}

func (h *BillingHandler) UpdateTransaction(c *fiber.Ctx) error {
	var body accounting_models.SoBillingTransaction
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	status, err := h.Service.UpdateTransaction(&body, at(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, body)
}

func (h *BillingHandler) DeleteTransaction(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	status, serviceErr := h.Service.DeleteTransaction(uint(id), at(c))
	if serviceErr != nil {
		return utils.RespondError(c, status, serviceErr.Error())
	}
	return utils.RespondSuccess(c, id)
}
