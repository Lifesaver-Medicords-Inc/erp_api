package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/sales_return_handlers"
	"github.com/pierceperado/smpc/services/sales_return_services"
)

// SalesReturnRoutes - Sales Return (SRT#), spec §5.13. Its own top-level
// group, same pattern as SalesOrderRoutes/DeliveryReceiptRoutes, rather than
// nested inside the older main.go route tree - this is brand-new
// functionality with no existing callers to preserve compatibility with.
func SalesReturnRoutes(app *fiber.App) {
	api := app.Group("/api/sales-returns")

	handler := sales_return_handlers.NewSalesReturnHandler(sales_return_services.NewSalesReturnService())

	api.Get("/", handler.GetSalesReturn)
	api.Get("/:id", handler.GetSalesReturnById)
	api.Post("/", handler.CreateSalesReturn)
	api.Post("/:id/approve", handler.ApproveSalesReturn)
}
