package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers/public_handlers"
	"github.com/pierceperado/smpc/handlers/sales_handlers"
	"github.com/pierceperado/smpc/handlers/setup_handlers"
	"github.com/pierceperado/smpc/initializers"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.MigrateDb()
}

func main() {
	// Fiber App
	app := fiber.New()

	// App Logger
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	// Api Endpoints
	api := app.Group("/api")
	{
		// Public Endpoints
		api.Post("/register", public_handlers.CreateAccount)
		api.Post("/login", public_handlers.LoginAccount)
		api.Post("/logout", public_handlers.LogoutAccount)
		api.Get("/hello", public_handlers.CheckHealth)

		// Protected Endpoints
		// api.Use(middlewares.RequireAuth)
		{
			// Setup Endpoints
			setupApi := api.Group("/setup")
			{
				// Brand Endpoints
				setupApi.Get("/brand", setup_handlers.GetBrands)
				setupApi.Get("/brand/:id", setup_handlers.GetBrand)
				setupApi.Post("/brand", setup_handlers.CreateBrand)
				setupApi.Put("/brand", setup_handlers.UpdateBrand)
				setupApi.Delete("/brand", setup_handlers.DeleteBrand)

				//Item Class Endpoints
				setupApi.Get("/item/class", setup_handlers.GetClasses)
				setupApi.Get("/item/class/:id", setup_handlers.GetClass)
				setupApi.Post("/item/class", setup_handlers.CreateClass)
				setupApi.Put("/item/class", setup_handlers.UpdateClass)
				setupApi.Delete("/item/class", setup_handlers.DeleteClass)

				//Item Name Endpoints
				setupApi.Get("/item/name", setup_handlers.GetNames)
				setupApi.Get("/item/name/:id", setup_handlers.GetName)
				setupApi.Post("/item/name", setup_handlers.CreateName)
				setupApi.Put("/item/name", setup_handlers.UpdateName)
				setupApi.Delete("/item/name", setup_handlers.DeleteName)

				//Item Name Endpoints
				setupApi.Get("/item/type", setup_handlers.GetTypes)
				setupApi.Get("/item/type/:id", setup_handlers.GetType)
				setupApi.Post("/item/type", setup_handlers.CreateType)
				setupApi.Put("/item/type", setup_handlers.UpdateType)
				setupApi.Delete("/item/type", setup_handlers.DeleteType)
				//Item Endpoints
				setupApi.Get("/item/", setup_handlers.GetClasses)
				setupApi.Get("/item/:id", setup_handlers.GetClass)
				setupApi.Post("/item", setup_handlers.CreateClass)
				setupApi.Put("/item", setup_handlers.UpdateClass)
				setupApi.Delete("/item", setup_handlers.DeleteClass)

				//Unit of Measurement Endpoint
				setupApi.Get("/unit_measurement", setup_handlers.GetUnitMeasurements)
				setupApi.Get("/unit_measurement:id", setup_handlers.GetUnitMeasurements)
				setupApi.Post("/unit_measurement", setup_handlers.CreateUnitMeasurement)
				setupApi.Put("/unit_measurement", setup_handlers.UpdateUnitMeasurement)
				setupApi.Delete("/unit_measurement", setup_handlers.DeleteUnitMeasurement)

				//Payment  Terms Endpoint
				setupApi.Get("/payment_terms", setup_handlers.GetPaymentTerms)
				setupApi.Get("/payment_terms:id", setup_handlers.GetPaymentTerm)
				setupApi.Post("/payment_terms", setup_handlers.CreatePaymentTerms)
				setupApi.Put("/payment_terms", setup_handlers.UpdatePaymentTerms)
				setupApi.Delete("/payment_terms", setup_handlers.DeletePaymentTerms)

			}

			// Sales Endpoints
			salesApi := api.Group("/sales")
			{
				// Quotation Endpoints
				salesApi.Get("/quotation", sales_handlers.GetQuotations)
				salesApi.Post("/quotation", sales_handlers.CreateQuotation)
				salesApi.Patch("/quotation", sales_handlers.UpdateQuotation)
				salesApi.Delete("/quotation", sales_handlers.DeleteQuotation)

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
		}
	}

	// Start Listen
	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
