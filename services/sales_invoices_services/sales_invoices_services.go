package sales_invoices_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type Body struct {
	accounting_models.SalesInvoice
	SalesInvoiceDetails []accounting_models.SalesInvoiceDetail `json:"sales_invoice_details"`
}

func GetSalesInvoices(conditions map[string]interface{}) (interface{}, int, error) {

	return "", 0, nil
}

func GetSalesInvoiceDocNo(conditions map[string]interface{}) (interface{}, int, error) {

	var response int
	if err := initializers.DB.Table("tbl_accounting_sales_invoice").
		Select("MAX(id)").
		Scan(&response).Error; err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get sales invoice doc_no")
	}

	return response, 0, nil
}

func CreateSalesInvoice(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesInvoice); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to insert sales invoice header")
	}

	at, ok := c.Locals("at").(models.At)
	userAt := utils.GetAtData(c, models.At{})
	at.AtUserId = userAt.AtUserId
	if !ok {
		at = models.At{}
	}

	parentChildAt := accounting_models.SalesInvoiceAt{RefId: body.ID, SalesInvoiceContent: body.SalesInvoiceContent, At: at}
	if err := services.DbInsert(tx, &parentChildAt); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	if err := CreateSalesInvoiceDetails(tx, &body.SalesInvoiceDetails, body.ID, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}
