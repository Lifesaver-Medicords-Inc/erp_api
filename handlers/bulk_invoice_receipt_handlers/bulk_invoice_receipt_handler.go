package bulk_invoice_receipt_handlers

import (
	"strconv"

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
	idStr := c.Query("id")
	seekIDStr := c.Query("seek_id")

	seekID := 0
	if seekIDStr != "" {
		parsed, err := strconv.Atoi(seekIDStr)
		if err != nil || parsed <= 0 {
			return utils.RespondError(c, fiber.StatusBadRequest, "invalid seek_id")
		}
		seekID = parsed
	}

	if idStr == "" {
		data, status, pagination, err := h.Service.GetBulkInvoiceReceipt(nil, 0, seekID)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data, pagination)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid id")
	}

	data, status, pagination, err := h.Service.GetBulkInvoiceReceipt(nil, id, seekID)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data, pagination)
}

func (h *BulkInvoiceReceiptHandler) GetBulkInvoiceReceiptSearch(c *fiber.Ctx) error {
	search := c.Query("search")

	var id int
	if idParam := c.Query("id"); idParam != "" {
		var err error
		id, err = strconv.Atoi(idParam)
		if err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "invalid id")
		}
	}

	data, status, pagination, err := h.Service.GetBulkInvoiceReceiptSearch(nil, search, id)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
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
