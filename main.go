package main

import (
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers/bpi_handlers"
	"github.com/pierceperado/smpc/handlers/position_handlers"
	"github.com/pierceperado/smpc/handlers/public_handlers"
	"github.com/pierceperado/smpc/handlers/purchasing_handlers"
	"github.com/pierceperado/smpc/handlers/sales_handlers"
	"github.com/pierceperado/smpc/handlers/sample_handlers"
	"github.com/pierceperado/smpc/handlers/setup_handlers"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectDb()
	// initializers.MigrateDb()
	initializers.InitRedis()
	initializers.InitWm()
	initializers.InitLogger()
}

func main() {

	// Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Static("/files", "./files")

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
		api.Post("/upload", public_handlers.ImageUpload)
		api.Post("/dfile", public_handlers.DeleteFile)
		api.Get("/vfile/:filename", public_handlers.ViewFile)

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

			//User Employee Endpoints
			api.Get("/employee_users/:employee_id", bpi_handlers.GetBpiUsers)

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

					// Material Endpoints
					itemApi.Get("/material", setup_handlers.GetMaterials)
					itemApi.Get("/material/:id", setup_handlers.GetMaterial)
					itemApi.Post("/material", setup_handlers.CreateMaterial)
					itemApi.Put("/material", setup_handlers.UpdateMaterial)
					itemApi.Delete("/material", setup_handlers.DeleteMaterial)

					// Name Endpoints
					itemApi.Get("/name", setup_handlers.GetNames)
					itemApi.Get("/name/:id", setup_handlers.GetName)
					itemApi.Post("/name", setup_handlers.CreateName)
					itemApi.Put("/name", setup_handlers.UpdateName)
					itemApi.Delete("/name", setup_handlers.DeleteName)

					// Pump Count Endpoints
					itemApi.Get("/pump_count", setup_handlers.GetPumpCounts)
					itemApi.Get("/pump_count/:id", setup_handlers.GetPumpCount)
					itemApi.Post("/pump_count", setup_handlers.CreatePumpCount)
					itemApi.Put("/pump_count", setup_handlers.UpdatePumpCount)
					itemApi.Delete("/pump_count", setup_handlers.DeletePumpCount)

					// Pump Type Endpoints
					itemApi.Get("/pump_type", setup_handlers.GetPumpTypes)
					itemApi.Get("/pump_type/:id", setup_handlers.GetPumpType)
					itemApi.Post("/pump_type", setup_handlers.CreatePumpType)
					itemApi.Put("/pump_type", setup_handlers.UpdatePumpType)
					itemApi.Delete("/pump_type", setup_handlers.DeletePumpType)

					// Type Endpoints
					itemApi.Get("/type", setup_handlers.GetTypes)
					itemApi.Get("/type/:id", setup_handlers.GetType)
					itemApi.Post("/type", setup_handlers.CreateType)
					itemApi.Put("/type", setup_handlers.UpdateType)
					itemApi.Delete("/type", setup_handlers.DeleteType)

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

				// Position Endpoints
				setupApi.Get("/position", position_handlers.GetPositions)
				setupApi.Get("/position:/id", position_handlers.GetPosition)
				setupApi.Post("/position", position_handlers.CreatePosition)
				setupApi.Put("/position", position_handlers.UpdatePosition)
				setupApi.Delete("/position", position_handlers.DeletePosition)

				//BOM Endpoints
				setupApi.Get("/bom", setup_handlers.GetSetupItemBoms)
				//setupApi.Get("/bom", setup_handlers.GetSetupItemBomss)

				setupApi.Get("/bom:/id", setup_handlers.GetSetupItemBom)
				setupApi.Post("/bom", setup_handlers.CreateSetupItemBom)
				setupApi.Put("/bom", setup_handlers.UpdateSetupItemBom)
				setupApi.Delete("/bom", setup_handlers.DeleteSetupItemBom)
				setupApi.Get("/bom/item_list", setup_handlers.GetBomItemList)
				setupApi.Get("/bom/parent_detail", setup_handlers.GetBomParentDetail)
				setupApi.Get("/bom/child_detail", setup_handlers.GetBomChildDetail)

				//BOM Detail Endpoints
				// setupApi.Get("/bom/detail", setup_handlers.GetSetupItemBomDetails)
				// setupApi.Get("/bom:/id", setup_handlers.GetSetupItemBomDetail)
				// setupApi.Post("/bom/detail", setup_handlers.CreateSetupItemBomDetail)
				// setupApi.Put("/bom/detail", setup_handlers.UpdateSetupItemBomDetail)
				// setupApi.Delete("/bom/detail", setup_handlers.DeleteSetupItemBomDetail)

				setupApi.Get("/project", setup_handlers.GetProjects)
				setupApi.Post("/project", setup_handlers.CreateProject)
				setupApi.Put("/project", setup_handlers.UpdateProject)

				// PROJECT TEMPLATES ENDPOINTS
				setupApi.Get("/templates", setup_handlers.GetProjectsTemplates)
				setupApi.Post("/templates", setup_handlers.CreateProjectTemplate)

				setupApi.Get("/boq", setup_handlers.GetItemBoqs)

				// =========================== ACCOUNTING ENDPOINTS =============================
				{
					// Chart Group Endpoints
					setupApi.Get("/book", setup_handlers.GetBooks)
					setupApi.Get("/book:/id", setup_handlers.GetChartGroup)
					setupApi.Post("/book", setup_handlers.CreateChartGroup)
					setupApi.Put("/book", setup_handlers.UpdateChartGroup)
					setupApi.Delete("/book", setup_handlers.DeleteChartGroup)

					// Chart Group Endpoints
					setupApi.Get("/chart_group", setup_handlers.GetChartGroups)
					setupApi.Get("/chart_group:/id", setup_handlers.GetChartGroup)
					setupApi.Post("/chart_group", setup_handlers.CreateChartGroup)
					setupApi.Put("/chart_group", setup_handlers.UpdateChartGroup)
					setupApi.Delete("/chart_group", setup_handlers.DeleteChartGroup)

					// Chart Class Endpoints
					setupApi.Get("/chart_class", setup_handlers.GetChartClasses)
					setupApi.Get("/chart_class:/id", setup_handlers.GetChartClass)
					setupApi.Post("/chart_class", setup_handlers.CreateChartClass)
					setupApi.Put("/chart_class", setup_handlers.UpdateChartClass)
					setupApi.Delete("/chart_class", setup_handlers.DeleteChartClass)

					// Chart Of Account Endpoints
					setupApi.Get("/chart_of_account", setup_handlers.GetChartOfAccounts)
					setupApi.Get("/chart_of_account:/id", setup_handlers.GetChartOfAccount)
					setupApi.Post("/chart_of_account", setup_handlers.CreateChartOfAccount)
					setupApi.Put("/chart_of_account", setup_handlers.UpdateChartOfAccount)
					setupApi.Delete("/chart_of_account", setup_handlers.DeleteChartOfAccount)

					// GetGeneralLedgerMappers
					setupApi.Get("/general_ledger", setup_handlers.GetGeneralLedgerMappers)
					setupApi.Get("/general_ledger:/id", setup_handlers.GetChartOfAccount)
					setupApi.Put("/general_ledger", setup_handlers.UpdateGeneralLedgerMappers)

					// Expanded Withholding Tax Endpoints
					setupApi.Get("/expanded_tax", setup_handlers.GetExpandedWithholdingTax)
					setupApi.Post("/expanded_tax", setup_handlers.CreateExpandedWithholdingTax)
					setupApi.Put("/expanded_tax", setup_handlers.UpdateExpandedWithholdingTax)
					setupApi.Delete("/expanded_tax", setup_handlers.DeleteExpandedWithholdingTax)

					// Expanded Withholding Tax Endpoints
					setupApi.Get("/final_tax", setup_handlers.GetFinalTax)
					setupApi.Post("/final_tax", setup_handlers.CreateFinalTax)
					setupApi.Put("/final_tax", setup_handlers.UpdateFinalTax)
					setupApi.Delete("/final_tax", setup_handlers.DeleteFinalTax)

				}
				// =========================== END =============================
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

				//salesApi.Post("/salescanvas", sales_handlers.CreateSalesCanvasSheet)
				//salesApi.Get("/salescanvas", sales_handlers.GetSalesCanvasView)

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
				salesApi.Put("/order", sales_handlers.UpdateOrder)
				salesApi.Delete("/order", sales_handlers.DeleteOrder)
				// Opportunity Endpointss
				salesApi.Get("/opportunity", sales_handlers.GetOpportunities)
				salesApi.Get("/opportunity/:id", sales_handlers.GetOpportunity)
				salesApi.Post("/opportunity", sales_handlers.CreateOpportunity)
				salesApi.Put("/opportunity", sales_handlers.UpdateOpportunity)

				//projects
				salesApi.Get("/projects", sales_handlers.GetSalesProject)
				salesApi.Post("/projects", sales_handlers.CreateSalesProject)
				salesApi.Post("/projects_tab", sales_handlers.CreateItemSetTab)
				salesApi.Put("/project_conditions", sales_handlers.UpdateProjectCondition)
				salesApi.Put("/project_contents", sales_handlers.UpdateProjectContent)

				//salesApi.Get("/projects_pumps", sales_handlers.GetItemPumps)

				// CRM Endpointss //test
				salesApi.Get("/crm", sales_handlers.GetCRMs)
				salesApi.Get("/crm/table", sales_handlers.GetCRMTable)
				salesApi.Get("/crm/:id", sales_handlers.GetCRM)
				salesApi.Post("/crm", sales_handlers.CreateCRM)
				salesApi.Put("/crm", sales_handlers.UpdateCRM)

				// Return Routes
				// sales_api.Get("/return", handlers.Register)
				// sales_api.Post("/return/create", handlers.Register)
				// sales_api.Patch("/return/update", handlers.Register)
				// sales_api.Delete("/return/delete", handlers.Register)
			}

			// Purchasing Endpoints
			purchasingApi := api.Group("/purchasing")
			{
				purchasingApi.Get("/purchase_requisition", purchasing_handlers.GetPRs)
				purchasingApi.Get("/purchase_requisition/:id", purchasing_handlers.GetPR)
				purchasingApi.Post("child/purchase_requisition", purchasing_handlers.CreatePRChild)
				purchasingApi.Post("/purchase_requisition", purchasing_handlers.CreatePR)
				purchasingApi.Put("/purchase_requisition", purchasing_handlers.UpdatePR)
				purchasingApi.Delete("/purchase_requisition", purchasing_handlers.DeletePR)
				purchasingApi.Delete("child/purchase_requisition", purchasing_handlers.DeletePROrderByID)

				// Purhcasing Redbox List
				purchasingApi.Get("/purchase_redbox_list", purchasing_handlers.GetPurchasingRedboxList)

				//Purchasing List
				purchasingApi.Get("/purchase_list", purchasing_handlers.GetSOPurchasingList)
				purchasingApi.Get("/purchase_list_supplier", purchasing_handlers.GetSOPurchasingListSupplier)
				purchasingApi.Get("/purchase_canvass_sheet", purchasing_handlers.GetPurchasingCanvassSheet)
				purchasingApi.Post("/purchase_canvass_sheet", purchasing_handlers.CreatePurchasingCanvassSheet)
				purchasingApi.Put("/purchase_canvass_sheet", purchasing_handlers.UpdatePurchasingCanvassSheet)
			}

			//Bpi Endpoints

			api.Get("/bpi/entity", bpi_handlers.GetBpiEntityRecords)
			api.Get("/bpi/list", bpi_handlers.GetBpiItemList)
			api.Post("/bpi", bpi_handlers.CreateBpi)
			api.Put("/bpi", bpi_handlers.UpdateBpi)
			api.Get("/bpi/:id", sales_handlers.GetBpi)
			//api.Get("/BpiSuppliers", sales_handlers.GetBpiSuppliers)

			api.Get("/bpi", bpi_handlers.GetBpis)

			// api.Get("/BpiSuppliers", sales_handlers.GetBpiSuppliers)
			//api.Delete("/bpi", sales_handlers.DeleteQuotation)

			// Websocket Endpoints
			ws := api.Group("/ws")
			{
				// Setup Endpoints
				setupApi := ws.Group("/setup")
				{
					// Item Endpoints
					itemApi := setupApi.Group("/item")
					{
						itemApi.Get("", websocket.New(setup_handlers.WsgetItems))
					}

					// Project Endpoints
					projectApi := setupApi.Group("/project")
					{
						projectApi.Get("", websocket.New(func(c *websocket.Conn) {
							services.HandleWs(c, setup_handlers.WsgetProjects)
						}))
					}
				}
				// Purchasing Endpoints
				purchasingApi := ws.Group("/purchasing")
				{
					redboxlistApi := purchasingApi.Group("/redboxlist")
					{
						redboxlistApi.Get("", websocket.New(func(c *websocket.Conn) {
							services.HandleWs(c, purchasing_handlers.WsgetRedboxList)
						}))
					}
				}
			}
		}
	}

	// Start Listen
	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
