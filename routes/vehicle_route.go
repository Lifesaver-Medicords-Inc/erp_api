package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func VehicleRoutes(app *fiber.App) {
	api := app.Group("/api/vehicles")

	vehicleService := adminservices.NewVehicleService()
	vehicleHandler := adminhandlers.NewVehicleHandler(vehicleService)
	api.Post("/", vehicleHandler.CreateVehicleHandler)
	api.Get("/:id", vehicleHandler.GetVehicleHandler)
	api.Get("/", vehicleHandler.GetVehiclesHandler)
	api.Put("/:id", vehicleHandler.UpdateVehicleHandler)
	api.Delete("/:id", vehicleHandler.DeleteVehicleHandler)

}
