package sales_invoice_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/sales_invoices_services"
	"github.com/pierceperado/smpc/utils"
)

func GetSalesInvoices(c *fiber.Ctx) error {

	data, status, err := sales_invoices_services.GetSalesInvoices(nil)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)

}

func GetSalesInvoiceDocNo(c *fiber.Ctx) error {
	data, status, err := sales_invoices_services.GetSalesInvoiceDocNo(nil)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
