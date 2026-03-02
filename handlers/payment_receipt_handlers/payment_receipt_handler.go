package payment_receipt_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/payment_receipt_services"
	"github.com/pierceperado/smpc/utils"
)

type PaymentReceiptHandler struct {
	Service *payment_receipt_services.PaymentReceiptService
}

func NewPaymentReceiptHandler(service *payment_receipt_services.PaymentReceiptService) *PaymentReceiptHandler {
	return &PaymentReceiptHandler{Service: service}
}

func (h *PaymentReceiptHandler) GetPaymentReceipt(c *fiber.Ctx) error {
	data, status, err := h.Service.GetPaymentReceipt(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PaymentReceiptHandler) CreatePaymentReceipt(c *fiber.Ctx) error {
	var body accounting_models.PaymentReceiptBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreatePaymentReceipt(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
