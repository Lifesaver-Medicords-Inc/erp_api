package ap_voucher_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/ap_voucher_services"
	"github.com/pierceperado/smpc/utils"
)

type ApVoucherHandler struct {
	Service *ap_voucher_services.ApVoucherService
}

func NewApVoucherHandler(service *ap_voucher_services.ApVoucherService) *ApVoucherHandler {
	return &ApVoucherHandler{Service: service}
}

func (h *ApVoucherHandler) GetApVoucher(c *fiber.Ctx) error {
	data, status, err := h.Service.GetApVoucher(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ApVoucherHandler) GetInvoiceView(c *fiber.Ctx) error {
	idParam := c.Params("supplier_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"SupplierId": idNum,
	}

	data, status, err := h.Service.GetInvoiceView(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ApVoucherHandler) CreateApVoucher(c *fiber.Ctx) error {
	var body accounting_models.ApVoucherBody
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateApVoucher(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
