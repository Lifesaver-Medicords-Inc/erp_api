package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers/bpi_handlers"
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
				salesApi.Get("/quotation", sales_handlers.GetSalesQuotations)
				//salesApi.Get("/quotation/:id", sales_handlers.GetSalesQuotation)
				//salesApi.Get("/quotation/:id", sales_handlers.GetBpi)
				//salesApi.Post("child/quotation", sales_handlers.CreateSalesQuotationChild)
				// POST for Parent
				salesApi.Post("/quotation", sales_handlers.CreateSalesQuotation)
				// salesApi.Put("/quotation", sales_handlers.UpdateSalesQuotation)
				// salesApi.Delete("/quotation", sales_handlers.DeleteSalesQuotation)

				salesApi.Get("/application", setup_handlers.GetApplications)
				salesApi.Get("/application/:id", setup_handlers.GetApplication)
				salesApi.Post("/application", setup_handlers.CreateApplication)
				salesApi.Put("/application", setup_handlers.UpdateApplication)
				salesApi.Delete("/application", setup_handlers.DeleteApplication)
				// Order Endpoints
				salesApi.Get("/order", sales_handlers.GetOrders)
				salesApi.Get("/order/:id", sales_handlers.GetOrder)
				salesApi.Post("child/order", sales_handlers.CreateOrderChild)
				salesApi.Post("/order", sales_handlers.CreateOrder)
				salesApi.Patch("/order", sales_handlers.UpdateOrder)
				salesApi.Delete("/order", sales_handlers.DeleteOrder)
				// Opportunity Endpointss
				salesApi.Get("/opportunity", sales_handlers.GetOpportunities)
				salesApi.Get("/opportunity/:id", sales_handlers.GetOpportunity)
				salesApi.Post("/opportunity", sales_handlers.CreateOpportunity)
				salesApi.Patch("/opportunity", sales_handlers.UpdateOpportunity)

				// Return Routes
				// sales_api.Get("/return", handlers.Register)
				// sales_api.Post("/return/create", handlers.Register)
				// sales_api.Patch("/return/update", handlers.Register)
				// sales_api.Delete("/return/delete", handlers.Register)

			}

			//Bpi Endpoints
			api.Get("/bpi", bpi_handlers.GetBpis)
			api.Get("/bpi/list", bpi_handlers.GetBpiItemList)
			api.Post("/bpi", bpi_handlers.CreateBpi)
			api.Get("/bpi/customers", sales_handlers.GetBpis)
			api.Get("/bpi/:id", sales_handlers.GetBpi)
			//api.Patch("/bpi", sales_handlers.UpdateQuotation)
			//api.Delete("/bpi", sales_handlers.DeleteQuotation)

		}
	}

	// Start Listen
	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
