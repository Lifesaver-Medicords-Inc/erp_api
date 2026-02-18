package payment_voucher_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/payment_voucher_services"
	"github.com/pierceperado/smpc/utils"
)

type PaymentVoucherHandler struct {
	Service *payment_voucher_services.PaymentVoucherService
}

func NewPaymentVoucherHandler(service *payment_voucher_services.PaymentVoucherService) *PaymentVoucherHandler {
	return &PaymentVoucherHandler{Service: service}
}

func (h *PaymentVoucherHandler) GetPaymentVoucher(c *fiber.Ctx) error {
	data, status, err := h.Service.GetPaymentVoucher(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PaymentVoucherHandler) GetSupplierAPVoucher(c *fiber.Ctx) error {
	idParam := c.Params("supplier_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"SupplierId": idNum,
	}

	data, status, err := h.Service.GetSupplierAPVoucher(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PaymentVoucherHandler) CreatePaymentVoucher(c *fiber.Ctx) error {
	var body accounting_models.PaymentVoucherBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreatePaymentVoucher(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
