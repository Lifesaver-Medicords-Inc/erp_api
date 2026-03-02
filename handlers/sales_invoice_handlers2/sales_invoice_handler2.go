package sales_invoice_handlers2

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/sales_invoice_services2"
	"github.com/pierceperado/smpc/utils"
)

type SalesInvoiceHandler struct {
	Service *sales_invoice_services2.SalesInvoiceService
}

func NewSalesInvoiceHandler(service *sales_invoice_services2.SalesInvoiceService) *SalesInvoiceHandler {
	return &SalesInvoiceHandler{Service: service}
}

func (h *SalesInvoiceHandler) GetExchangeRate(c *fiber.Ctx) error {
	baseCode := c.Params("base_code")

	data, status, err := h.Service.GetExchangeRate(baseCode)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesInvoiceHandler) GetCustomer(c *fiber.Ctx) error {
	data, status, err := h.Service.GetCustomer(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesInvoiceHandler) GetCustomerSO(c *fiber.Ctx) error {
	idParam := c.Params("customer_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"CustomerId": idNum,
	}

	data, status, err := h.Service.GetCustomerSO(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesInvoiceHandler) GetSalesInvoice(c *fiber.Ctx) error {
	data, status, err := h.Service.GetSalesInvoice(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesInvoiceHandler) CreateSalesInvoice(c *fiber.Ctx) error {
	var body accounting_models.SalesInvoice2Body

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateSalesInvoice(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
