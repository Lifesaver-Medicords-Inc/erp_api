package bulk_invoice_receipt_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/bulk_invoice_receipt_services"
	"github.com/pierceperado/smpc/utils"
)

type BulkInvoiceReceiptHandler struct {
	Service *bulk_invoice_receipt_services.BulkInvoiceReceiptService
}

func NewBulkInvoiceReceiptHandler(service *bulk_invoice_receipt_services.BulkInvoiceReceiptService) *BulkInvoiceReceiptHandler {
	return &BulkInvoiceReceiptHandler{Service: service}
}

func (h *BulkInvoiceReceiptHandler) GetBulkInvoiceReceipt(c *fiber.Ctx) error {
	data, status, err := h.Service.GetBulkInvoiceReceipt(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *BulkInvoiceReceiptHandler) CreateBulkInvoiceReceipt(c *fiber.Ctx) error {
	var body accounting_models.BulkInvoiceReceiptBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateBulkInvoiceReceipt(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
