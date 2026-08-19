package routes

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {
	// Root API group
	api := app.Group("/api")

	// User routes
	UserRoutes(app)

	// Position Routes
	PositionRoutes(app)

	// Position access routes
	PositionAccessRoutes(app)

	// Access module catalog routes (read-only list of every grantable
	// screen/action, for the Access Control screen's tree)
	AccessModuleRoutes(app)

	// Permission routes
	PermissionRoutes(app)

	// Vehicle routes
	VehicleRoutes(app)

	// warehouse routes
	WarehouseRoutes(app)

	// File routes
	VehicleFileRoutes(app)
	ReceiptUploadRoutes(app)

	// Company Routes
	CompanyRoutes(app)

	// Currency Routes
	CurrencyRoutes(app)

	CalendarScheduleRoutes(app)

	DeliveryReceiptRoutes(app)

	ItemReleaseRoutes(app)

	SalesOrderRoutes(app)

	AccountingRoutes(api)

	EngineeringRoutes(api)

	InventoryRoutes(api)

	AdminRoutes(api)
}
