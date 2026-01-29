package invoice_receipt_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/invoice_receipt_services"
	"github.com/pierceperado/smpc/utils"
)

type InvoiceReceiptHandler struct {
	Service *invoice_receipt_services.InvoiceReceiptService
}

func NewInvoiceReceiptHandler(service *invoice_receipt_services.InvoiceReceiptService) *InvoiceReceiptHandler {
	return &InvoiceReceiptHandler{Service: service}
}

func (h *InvoiceReceiptHandler) GetInvoiceReceipt(c *fiber.Ctx) error {
	data, status, err := h.Service.GetInvoiceReceipt(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *InvoiceReceiptHandler) GetSupplierPO(c *fiber.Ctx) error {
	idParam := c.Params("supplier_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"SupplierId": idNum,
	}

	data, status, err := h.Service.GetSupplierPO(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *InvoiceReceiptHandler) GetTaxView(c *fiber.Ctx) error {
	data, status, err := h.Service.GetTaxView(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *InvoiceReceiptHandler) GetSupplierTradeView(c *fiber.Ctx) error {
	data, status, err := h.Service.GetSupplierTradeView(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *InvoiceReceiptHandler) CreateInvoiceReceipt(c *fiber.Ctx) error {
	var body accounting_models.InvoiceReceiptBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateInvoiceReceipt(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
