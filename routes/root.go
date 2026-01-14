package routes

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {

	// Root API group
	api := app.Group("/api")

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

	//Company Routes
	CompanyRoutes(app)

	//Currency Routes
	CurrencyRoutes(app)

	CalendarScheduleRoutes(app)

	DeliveryReceiptRoutes(app)

	ItemReleaseRoutes(app)

	SalesOrderRoutes(app)

	AccountingRoutes(api)
}
