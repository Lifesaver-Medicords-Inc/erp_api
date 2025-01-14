package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers"
	sales_quotation_handler "github.com/pierceperado/smpc/handlers/sales"
	setup_brand_handler "github.com/pierceperado/smpc/handlers/setup"
	"github.com/pierceperado/smpc/initializers"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.MigrateDb()
}

func main() {
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	api := app.Group("/api")
	{
		api.Post("/register", handlers.Register)
		api.Post("/login", handlers.Login)
		api.Post("/logout", handlers.Logout)
	}

	// SALES ROUTES
	sales_api := app.Group("/api/sales")
	{
		// Quotation Routes
		sales_api.Get("/quotation", sales_quotation_handler.Get)
		sales_api.Post("/quotation/create", sales_quotation_handler.Create)
		sales_api.Patch("/quotation/update", sales_quotation_handler.Update)
		sales_api.Delete("/quotation/delete", sales_quotation_handler.Delete)

		// Order Routes
		// sales_api.Get("/order", handlers.Register)
		// sales_api.Post("/order/create", handlers.Register)
		// sales_api.Patch("/order/update", handlers.Register)
		// sales_api.Delete("/order/delete", handlers.Register)

		// Return Routes
		// sales_api.Get("/return", handlers.Register)
		// sales_api.Post("/return/create", handlers.Register)
		// sales_api.Patch("/return/update", handlers.Register)
		// sales_api.Delete("/return/delete", handlers.Register)

	}

	setup_api := app.Group("/api/setup")
	{

		setup_api.Get("/brand", setup_brand_handler.Get)
		setup_api.Post("/brand/create", setup_brand_handler.Create)
		setup_api.Patch("/brand/update", setup_brand_handler.Update)
		setup_api.Delete("/brand/delete", setup_brand_handler.Delete)

	}

	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
