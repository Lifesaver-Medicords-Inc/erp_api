package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func WarehouseRoutes(app *fiber.App) {
	api := app.Group("/api/warehouses")

	warehouseService := adminservices.NewWarehouseService()
	warehouseHandler := adminhandlers.NewWarehouseHandler(warehouseService)
	api.Post("/", warehouseHandler.CreateWarehouse)
	api.Get("/", warehouseHandler.GetWarehouses)
	api.Get("/:id", warehouseHandler.GetWarehouse)

}
