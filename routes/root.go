package routes

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {
	//User routes
	UserRoutes(app)

	//Position Routes
	PositionRoutes(app)

	//Position access routes
	PositionAccessRoutes(app)

	//Permission routes
	PermissionRoutes(app)

	//Vehicle routes
	VehicleRoutes(app)

	//warehouse routes
	WarehouseRoutes(app)

	//File routes
	VehicleFileRoutes(app)
}
