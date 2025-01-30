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
	initializers.InitRedis()
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

				itemApi := setupApi.Group("/item")
				{
					// Brand Endpoints
					itemApi.Get("/brand", setup_handlers.GetBrands)
					itemApi.Get("/brand/:id", setup_handlers.GetBrand)
					itemApi.Post("/brand", setup_handlers.CreateBrand)
					itemApi.Put("/brand", setup_handlers.UpdateBrand)
					itemApi.Delete("/brand", setup_handlers.DeleteBrand)

					// Class Endpoints
					itemApi.Get("/class", setup_handlers.GetClasses)
					itemApi.Get("/class/:id", setup_handlers.GetClass)
					itemApi.Post("/class", setup_handlers.CreateClass)
					itemApi.Put("/class", setup_handlers.UpdateClass)
					itemApi.Delete("/class", setup_handlers.DeleteClass)

					// Name Endpoints
					itemApi.Get("/name", setup_handlers.GetNames)
					itemApi.Get("/name/:id", setup_handlers.GetName)
					itemApi.Post("/name", setup_handlers.CreateName)
					itemApi.Put("/name", setup_handlers.UpdateName)
					itemApi.Delete("/name", setup_handlers.DeleteName)

					// Type Endpoints
					itemApi.Get("/type", setup_handlers.GetTypes)
					itemApi.Get("/type/:id", setup_handlers.GetType)
					itemApi.Post("/type", setup_handlers.CreateType)
					itemApi.Put("/type", setup_handlers.UpdateType)
					itemApi.Delete("/type", setup_handlers.DeleteType)

					// Model Endpoints
					itemApi.Get("/model", setup_handlers.GetModels)
					itemApi.Get("/model/:id", setup_handlers.GetModel)
					itemApi.Post("/model", setup_handlers.CreateModel)
					itemApi.Put("/model", setup_handlers.UpdateModel)
					itemApi.Delete("/model", setup_handlers.DeleteModel)

					// Item Endpoints
					itemApi.Get("", setup_handlers.GetItems)
					itemApi.Get("/:id", setup_handlers.GetItem)
					itemApi.Post("", setup_handlers.CreateItem)
					itemApi.Put("", setup_handlers.UpdateItem)
					itemApi.Delete("", setup_handlers.DeleteItem)
				}

				// Unit Measurement Endpoints
				setupApi.Get("/unit_measurement", setup_handlers.GetUnitMeasurements)
				setupApi.Get("/unit_measurement:id", setup_handlers.GetUnitMeasurements)
				setupApi.Post("/unit_measurement", setup_handlers.CreateUnitMeasurement)
				setupApi.Put("/unit_measurement", setup_handlers.UpdateUnitMeasurement)
				setupApi.Delete("/unit_measurement", setup_handlers.DeleteUnitMeasurement)

				// Payment Terms Endpoints
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
				salesApi.Get("/quotation", sales_handlers.GetSalesQuotations)
				salesApi.Get("/quotation/:id", sales_handlers.GetSalesQuotation)
				// POST for child
				salesApi.Post("/quotation", sales_handlers.CreateSalesQuotationChild)
				// POST for Parent
				salesApi.Post("/quotation", sales_handlers.CreateSalesQuotation)
				salesApi.Put("/quotation", sales_handlers.UpdateSalesQuotation)
				salesApi.Delete("/quotation", sales_handlers.DeleteSalesQuotation)

				salesApi.Get("/application", setup_handlers.GetApplications)
				salesApi.Get("/application/:id", setup_handlers.GetApplication)
				salesApi.Post("/application", setup_handlers.CreateApplication)
				salesApi.Put("/application", setup_handlers.UpdateApplication)
				salesApi.Delete("/application", setup_handlers.DeleteApplication)

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
