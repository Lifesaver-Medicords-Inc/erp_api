package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func SalesOrderRoutes(app *fiber.App) {
	api := app.Group("/api/sales-orders")

	salesOrderService := dispatching_services.NewSalesOrderService()
	salesOrderHandler := dispatching_handlers.NewSalesOrderHandler(salesOrderService)
	api.Get("/", salesOrderHandler.GetSalesOrdersHandler)
	api.Get("/:id", salesOrderHandler.GetSalesOrderHandler)
	api.Post("/", salesOrderHandler.CreateSalesOrderHandler)
	api.Put("/:id", salesOrderHandler.UpdateSalesOrderHandler)
	api.Delete("/:id", salesOrderHandler.DeleteSalesOrderHandler)

}
