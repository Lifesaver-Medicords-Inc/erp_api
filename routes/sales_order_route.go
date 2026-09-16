package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	"github.com/pierceperado/smpc/handlers/sales_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/services/sales_services"
)

func SalesOrderRoutes(app *fiber.App) {
	api := app.Group("/api/sales-orders")

	salesOrderService := dispatching_services.NewSalesOrderService()
	salesOrderHandler := dispatching_handlers.NewSalesOrderHandler(salesOrderService)
	api.Get("/", salesOrderHandler.GetSalesOrdersHandler)
	api.Get("/:id", salesOrderHandler.GetSalesOrderHandler)
	//api.Post("/", salesOrderHandler.CreateSalesOrderHandler)
	// api.Put("/:id", salesOrderHandler.UpdateSalesOrderHandler)
	// api.Delete("/:id", salesOrderHandler.DeleteSalesOrderHandler)

	setupOrderChargeRoutes(api)
}

// Sales Order cancellation charges (spec 5.4, 8.15). Cancelling an approved SO is
// a two-stage act: /charges raises it and parks the order at FOR REVIEW, and only
// /charges/decision moves it to CANCELLED - which is why raising and deciding are
// separate endpoints rather than one "cancel" call.
func setupOrderChargeRoutes(api fiber.Router) {
	handler := sales_handlers.NewOrderChargeHandler(sales_services.NewOrderChargeService())

	// The two percentages the charges modal autofills with (4.5.6).
	api.Get("/charges/defaults", handler.Defaults)
	// The approver's queue.
	api.Get("/charges/pending", handler.Pending)
	// What PROCEED would charge at the percentages currently typed in.
	api.Get("/:id/charges/preview", handler.Preview)
	// The charge record in force for one order.
	api.Get("/:id/charges", handler.Get)
	// PROCEED.
	api.Post("/charges", handler.Raise)
	// The Sales Manager's or CBDO's approve/reject.
	api.Post("/charges/decision", handler.Decide)
}
