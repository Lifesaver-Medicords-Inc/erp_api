package sales_invoices_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
)

// type SalesInvoiceDocNo struct {
// 	DocNo uint `json:'doc_no'`
// }

func GetSalesInvoices(conditions map[string]interface{}) (interface{}, int, error) {

	return "", 0, nil
}

func GetSalesInvoiceDocNo(conditions map[string]interface{}) (interface{}, int, error) {

	var response int
	if err := initializers.DB.Table("tbl_bpi").
		Select("MAX(id)").
		Scan(&response).Error; err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get sales invoice doc_no")
	}

	return response, 0, nil
}
