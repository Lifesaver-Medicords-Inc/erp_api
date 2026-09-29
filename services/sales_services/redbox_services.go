package sales_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// GetSalesRedBoxQuotations backs the Sales red box's quote cards: one row per open quote, both
// Quick and Project (vw_get_sales_redbox_quotations). Uncached - a quote must show on the
// rail as soon as it is saved, and the view is a single pass over the quotation headers.
func GetSalesRedBoxQuotations() ([]models.SalesRedBoxQuotationView, int, error) {
	rows := []models.SalesRedBoxQuotationView{}

	if err := services.DbGetNoCache(&rows, nil); err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed getting red box quotations")
	}

	return rows, 0, nil
}
