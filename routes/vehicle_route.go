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
	api.Post("/", vehicleHandler.CreateVehicle)
	api.Get("/:id", vehicleHandler.GetVehicle)
	api.Get("/", vehicleHandler.GetVehicles)
	api.Put("/:id", vehicleHandler.UpdateVehicle)
	api.Delete("/:id", vehicleHandler.DeleteVehicle)

}
