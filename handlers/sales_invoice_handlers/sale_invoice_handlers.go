package sales_invoice_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
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
func CreateSalesInvoice(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_invoices_services.CreateSalesInvoice(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())

	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transactions")
	}
	return utils.RespondSuccess(c, data)
}
