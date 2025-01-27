package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers/public_handlers"
	"github.com/pierceperado/smpc/handlers/sales_handlers"
	"github.com/pierceperado/smpc/handlers/sample_handlers"
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
			// Sample Endpoints
			sampleApi := api.Group("/sample")
			{
				sampleApi.Get("/parent", sample_handlers.GetParents)
				sampleApi.Get("/parent/:id", sample_handlers.GetParent)
				sampleApi.Post("/parent", sample_handlers.CreateParent)
				sampleApi.Put("/parent", sample_handlers.UpdateParent)
				sampleApi.Delete("/parent", sample_handlers.DeleteParent)
			}

			// Setup Endpoints
			setupApi := api.Group("/setup")
			{
				// Item Endpoints
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
				}

				// Unit Measurement Endpoints
				setupApi.Get("/unit_measurement", setup_handlers.GetUnitMeasurements)
				setupApi.Get("/unit_measurement:/id", setup_handlers.GetUnitMeasurement)
				setupApi.Post("/unit_measurement", setup_handlers.CreateUnitMeasurement)
				setupApi.Put("/unit_measurement", setup_handlers.UpdateUnitMeasurement)
				setupApi.Delete("/unit_measurement", setup_handlers.DeleteUnitMeasurement)

				// Payment Terms Endpoints
				setupApi.Get("/payment_terms", setup_handlers.GetPaymentTerms)
				setupApi.Get("/payment_terms:/id", setup_handlers.GetPaymentTerm)
				setupApi.Post("/payment_terms", setup_handlers.CreatePaymentTerms)
				setupApi.Put("/payment_terms", setup_handlers.UpdatePaymentTerms)
				setupApi.Delete("/payment_terms", setup_handlers.DeletePaymentTerms)

				// Ship Type Endpoints
				setupApi.Get("/shiptype", setup_handlers.GetShipTypes)
				setupApi.Get("/shiptype/:id", setup_handlers.GetShipType)
				setupApi.Post("/shiptype", setup_handlers.CreateShipType)
				setupApi.Put("/shiptype", setup_handlers.UpdateShipType)
				setupApi.Delete("/shiptype", setup_handlers.DeleteShipType)

				// Social Media Endpoints
				setupApi.Get("/social", setup_handlers.GetSocial)
				setupApi.Get("/social:/id", setup_handlers.GetSocialMedia)
				setupApi.Post("/social", setup_handlers.CreateSocial)
				setupApi.Put("/social", setup_handlers.UpdateSocial)
				setupApi.Delete("/social", setup_handlers.DeleteSocial)

				//Industries Endpoints
				setupApi.Get("/industries", setup_handlers.GetIndustries)
				setupApi.Get("/industries:/id", setup_handlers.GetIndustry)
				setupApi.Post("/industries", setup_handlers.CreateIndustry)
				setupApi.Put("/industries", setup_handlers.UpdateIndustry)
				setupApi.Delete("/industries", setup_handlers.DeleteIndustry)

				//Entity Type Endpoints
				setupApi.Get("/entity", setup_handlers.GetEntities)
				setupApi.Get("/entity:/id", setup_handlers.GetEntity)
				setupApi.Post("/entity", setup_handlers.CreateEntity)
				setupApi.Put("/entity", setup_handlers.UpdateEntity)
				setupApi.Delete("/entity", setup_handlers.DeleteEntity)
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
			// Bpi Endpoints

		}
	}

	// Start Listen
	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
