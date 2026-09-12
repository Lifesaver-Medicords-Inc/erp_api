package sales_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/sales_services"
	"github.com/pierceperado/smpc/utils"
)

// Sales Order cancellation charges (spec 5.4, 8.15).
type OrderChargeHandler struct {
	Service *sales_services.OrderChargeService
}

func NewOrderChargeHandler(service *sales_services.OrderChargeService) *OrderChargeHandler {
	return &OrderChargeHandler{Service: service}
}

func chargeAt(c *fiber.Ctx) models.At {
	value, ok := c.Locals("at").(models.At)
	if !ok {
		return models.At{}
	}
	return value
}

// The two percentages the charges modal autofills with (4.5.6).
func (h *OrderChargeHandler) Defaults(c *fiber.Ctx) error {
	restocking, cancellation := h.Service.Defaults()
	return utils.RespondSuccess(c, fiber.Map{
		"restocking_fee_percent":   restocking,
		"cancellation_fee_percent": cancellation,
	})
}

// What PROCEED would charge, at the percentages currently typed in the modal.
// Both percentages are read from the query string so a 0 is explicit - 0 means
// the fee is declined (5.4) and must never fall back to a default.
func (h *OrderChargeHandler) Preview(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	restocking, _ := strconv.ParseFloat(c.Query("restocking_fee_percent", "0"), 64)
	cancellation, _ := strconv.ParseFloat(c.Query("cancellation_fee_percent", "0"), 64)

	preview, status, serviceErr := h.Service.Compute(uint(id), restocking, cancellation)
	if serviceErr != nil {
		return utils.RespondError(c, status, serviceErr.Error())
	}
	return utils.RespondSuccess(c, preview)
}

func (h *OrderChargeHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	charge, status, serviceErr := h.Service.Get(uint(id))
	if serviceErr != nil {
		return utils.RespondError(c, status, serviceErr.Error())
	}
	return utils.RespondSuccess(c, charge)
}

func (h *OrderChargeHandler) Pending(c *fiber.Ctx) error {
	rows, status, err := h.Service.Pending()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, rows)
}

// PROCEED on the charges modal.
func (h *OrderChargeHandler) Raise(c *fiber.Ctx) error {
	var body models.SalesOrderCharge
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	status, err := h.Service.Raise(&body, chargeAt(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, body)
}

type chargeDecisionBody struct {
	ID           uint   `json:"id"`
	Approve      bool   `json:"approve"`
	ReviewedBy   string `json:"reviewed_by"`
	ReviewedById uint   `json:"reviewed_by_id"`
	ReviewedDate string `json:"reviewed_date"`
}

// The Sales Manager's or CBDO's decision. The same access code that gates
// cancelling an order gates deciding one - a sales executive may raise a
// cancellation but never approve their own.
func (h *OrderChargeHandler) Decide(c *fiber.Ctx) error {
	var body chargeDecisionBody
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}
	if body.ID == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "id is required")
	}

	userId := body.ReviewedById
	allowed, err := sales_services.UserHasOrderAccess(userId, sales_services.OrderCancelAccessCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed checking access")
	}
	if !allowed {
		return utils.RespondError(c, fiber.StatusForbidden,
			"only the Sales Manager or the CBDO may approve a cancellation (spec 5.4)")
	}

	status, serviceErr := h.Service.Approve(body.ID, body.Approve, body.ReviewedBy, body.ReviewedById,
		body.ReviewedDate, chargeAt(c))
	if serviceErr != nil {
		return utils.RespondError(c, status, serviceErr.Error())
	}
	return utils.RespondSuccess(c, body)
}
