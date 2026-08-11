package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers/bpi_handlers"
	"github.com/pierceperado/smpc/handlers/engineering_handlers"
	"github.com/pierceperado/smpc/handlers/item_request_handlers"
	"github.com/pierceperado/smpc/handlers/pick_activity_handlers"
	"github.com/pierceperado/smpc/handlers/position_handlers"
	"github.com/pierceperado/smpc/handlers/public_handlers"
	"github.com/pierceperado/smpc/handlers/purchasing_handlers"
	"github.com/pierceperado/smpc/handlers/sales_handlers"
	"github.com/pierceperado/smpc/handlers/sample_handlers"
	"github.com/pierceperado/smpc/handlers/setup_handlers"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/migrations"
	"github.com/pierceperado/smpc/routes"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.MigrateAll()
	initializers.MigrateModel("accounting")
	// Only adds the new z_tbl_inv_item_stocks_at.remarks column (see migrateInventoryWarehouse) -
	// migrateInventoryWarehouse() itself is otherwise all commented out, so this doesn't touch
	// any of the other legacy inventory models still sitting in that block.
	initializers.MigrateModel("inventory")
	initializers.InitRedis()
	initializers.InitWm()
	initializers.InitWm2()
	initializers.InitProjectWM()
	initializers.InitWmJobOrder()
	initializers.InitWmQuotation()
	initializers.InitLogger()
	migrations.RunSQLMigrations()
	startStockReservationSweep()
}

// startStockReservationSweep periodically deletes expired rows from
// tbl_inv_stock_reservations (see ExpireStockReservations) so a quotation's soft hold
// on stock doesn't outlive its own ValidUntil. There's no existing job scheduler in
// this app, so this is a plain goroutine + ticker - deliberately placed here rather
// than in the initializers package, since item_stock_services already imports
// initializers and putting it there would create an import cycle.
func startStockReservationSweep() {
	stockService := item_stock_services.NewItemStockService()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			tx := initializers.DB.Begin()
			if tx.Error != nil {
				fmt.Println("stock reservation sweep: failed starting transaction:", tx.Error)
				continue
			}

			count, err := stockService.ExpireStockReservations(tx)
			if err != nil {
				fmt.Println("stock reservation sweep: failed:", err)
				tx.Rollback()
				continue
			}

			if err := tx.Commit().Error; err != nil {
				fmt.Println("stock reservation sweep: failed committing:", err)
				continue
			}

			if count > 0 {
				fmt.Println("stock reservation sweep: released", count, "expired reservation(s)")
			}
		}
	}()
}

func main() {
	app := SetupApp()

	// Start Listen
	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}

func SetupApp() *fiber.App {
	// Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	// Rate Limiter
	// app.Use(limiter.New(limiter.Config{
	// 	Max:        20,
	// 	Expiration: 1 * time.Second,
	// }))

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
				invApi := setupApi.Group("/inv")
				{
					//Inventory Tracker endpoints
					invApi.Get("/tracker", setup_handlers.GetInvTracker)
					invApi.Get("/warehouse_name", setup_handlers.GetInvWarehouseName)
					invApi.Post("/tracker", setup_handlers.CreateInvTracker)
					invApi.Put("/tracker", setup_handlers.UpdateInvTracker)
					invApi.Delete("/tracker", setup_handlers.DeleteInvTracker)

					//Inventory Logbook endpoints
					invApi.Get("/logbook", setup_handlers.GetInvLogbook)
				}

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

					// Item Request Endpoints
					itemApi.Get("/request", item_request_handlers.GetItemRequest)
					itemApi.Get("/all_item", item_request_handlers.GetAllItemList)
					itemApi.Get("/all_binloc/:itemId", item_request_handlers.GetAllBinLocation)
					itemApi.Get("/all_user", item_request_handlers.GetUserList)
					itemApi.Get("/so_doc", item_request_handlers.GetSalesOrderIR)
					itemApi.Post("/request", item_request_handlers.CreateItemRequest)
					itemApi.Put("/request", item_request_handlers.UpdateItemRequest)
					itemApi.Delete("/request", item_request_handlers.DeleteItemRequest)

					// Item Valuation Method Endpoints
					itemApi.Get("/valuation_method", setup_handlers.GetValuationMethods)
					itemApi.Get("/valuation_method/:id", setup_handlers.GetValuationMethod)
					itemApi.Post("/valuation_method", setup_handlers.CreateValuationMethod)
					itemApi.Put("/valuation_method", setup_handlers.UpdateValuationMethod)
					itemApi.Delete("/valuation_method", setup_handlers.DeleteValuationMethod)

					// Item Endpoints
					itemApi.Get("", setup_handlers.GetItems)
					itemApi.Get("/:id", setup_handlers.GetItem)
					itemApi.Post("", setup_handlers.CreateItem)
					itemApi.Put("", setup_handlers.UpdateItem)
					itemApi.Delete("", setup_handlers.DeleteItem)
					itemApi.Post("/migrate", setup_handlers.CreateItemByMigration)
				}

				//Pick Activity
				pickActivityApi := setupApi.Group("/pickAct")
				{
					pickActivityApi.Get("/binloc", pick_activity_handlers.GetBinLocation)
					pickActivityApi.Get("/list", pick_activity_handlers.GetPickActivity)
					pickActivityApi.Get("/salesOrder", pick_activity_handlers.GetSalesOrderPA)
					pickActivityApi.Post("/list", pick_activity_handlers.CreatePickActivity)
					pickActivityApi.Put("/list", pick_activity_handlers.UpdatePickActivity)
					pickActivityApi.Delete("/list", pick_activity_handlers.DeletePickActivity)
				}

				//warehouse
				warehouseApi := setupApi.Group("/warehouse")
				{
					warehouseApi.Get("/manager", setup_handlers.GetWarehouseManagers) //for warehouse manager
					//warehouse name
					warehouseApi.Get("/name", setup_handlers.GetWarehouseNames)
					warehouseApi.Get("/name/:id", setup_handlers.GetWarehouseName)
					warehouseApi.Post("/name", setup_handlers.CreateWarehouseName)
					warehouseApi.Put("/name", setup_handlers.UpdateWarehouseName)
					warehouseApi.Delete("/name", setup_handlers.DeleteWarehouseName)

					//warehouse usetypes
					warehouseApi.Get("/usetype", setup_handlers.GetUseTypes)
					warehouseApi.Get("/usetype/:id", setup_handlers.GetUseType)
					warehouseApi.Post("/usetype", setup_handlers.CreateUseType)
					warehouseApi.Put("/usetype", setup_handlers.UpdateUseType)
					warehouseApi.Delete("/usetype", setup_handlers.DeleteUseType)

					//no endpoint (saved as a package with parent(warehousename))
					//warehouse address and for warehouse area
					// warehouseApi.Get("/address", setup_handlers.GetWarehouseAddresses)
					// warehouseApi.Get("/address/:id", setup_handlers.GetWarehouseAddress)
					// warehouseApi.Post("/address", setup_handlers.CreateWarehouseAddress)
					// warehouseApi.Put("/address", setup_handlers.UpdateWarehouseAddress)
					// warehouseApi.Delete("/address", setup_handlers.DeleteWarehouseAddress)

					//Warehouse Areas (used for separate saving, pagsamantagal)
					warehouseApi.Get("/area", setup_handlers.GetWarehouseAreasRow)
					warehouseApi.Get("/area/:id", setup_handlers.GetWarehouseAreaRow)
					warehouseApi.Post("/area", setup_handlers.CreateWarehouseAreaRow)
					warehouseApi.Put("/area", setup_handlers.UpdateWarehouseAreaRow)
					warehouseApi.Delete("/area", setup_handlers.DeleteWarehouseAreaRow)
				}

				reports2Api := setupApi.Group("/report2")
				{
					//receiving report
					reports2Api.Get("/receiving2", setup_handlers.GetReceivingReports2)
					reports2Api.Get("/purchase_filter", setup_handlers.GetPurchaseOrderView)
					reports2Api.Get("/receiving2/:id", setup_handlers.GetReceivingReport2)
					reports2Api.Post("/receiving2", setup_handlers.CreateReceivingReport2)
					reports2Api.Put("/receiving2", setup_handlers.UpdateReceivingReport2)
					reports2Api.Delete("/receiving2", setup_handlers.DeleteReceivingReport2)
					//receiving report detail
					reports2Api.Delete("/receiving_details2", setup_handlers.DeleteReceivingReportDetailsRow2)
					reports2Api.Get("/purchase_order/:po_id", setup_handlers.GetPurchaseOrderDetails)
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
				setupApi.Get("/all_bom/item_list", setup_handlers.GetAllBomItemList)
				setupApi.Get("/bom/parent_detail", setup_handlers.GetBomParentDetail)
				setupApi.Get("/bom/child_detail", setup_handlers.GetBomChildDetail)

				//BOM Detail Endpoints
				// setupApi.Get("/bom/detail", setup_handlers.GetSetupItemBomDetails)
				// setupApi.Get("/bom:/id", setup_handlers.GetSetupItemBomDetail)
				// setupApi.Post("/bom/detail", setup_handlers.CreateSetupItemBomDetail)
				// setupApi.Put("/bom/detail", setup_handlers.UpdateSetupItemBomDetail)
				// setupApi.Delete("/bom/detail", setup_handlers.DeleteSetupItemBomDetail)

				// test project
				setupApi.Get("/project", setup_handlers.GetProjects)
				setupApi.Post("/project", setup_handlers.CreateProject)
				setupApi.Put("/project", setup_handlers.UpdateProject)

				// PROJECT TEMPLATES ENDPOINTS
				setupApi.Get("/templates", setup_handlers.GetProjectsTemplates)
				setupApi.Post("/templates", setup_handlers.CreateProjectTemplate)
				setupApi.Put("/templates", setup_handlers.UpdateProjectTemplate)
				setupApi.Delete("/templates", setup_handlers.DeleteProjectTemplate)

				setupApi.Get("/boq", setup_handlers.GetItemBoqs)

				// =========================== ACCOUNTING ENDPOINTS SETUP =============================
				{
					// Chart Book Endpoints
					setupApi.Get("/book", setup_handlers.GetBooks)
					setupApi.Get("/book:/id", setup_handlers.GetChartGroup)
					setupApi.Post("/book", setup_handlers.CreateChartGroup)
					setupApi.Put("/book", setup_handlers.UpdateChartGroup)
					setupApi.Delete("/book", setup_handlers.DeleteChartGroup)

					// GetGeneralLedgerMappers
					setupApi.Get("/general_ledger", setup_handlers.GetGeneralLedgerMappers)
					//	setupApi.Get("/general_ledger:/id", setup_handlers.GetChartOfAccount)
					setupApi.Put("/general_ledger", setup_handlers.UpdateGeneralLedgerMappers)

					// Expanded Withholding Tax Endpoints
					setupApi.Get("/expanded_tax", setup_handlers.GetExpandedWithholdingTax)
					setupApi.Post("/expanded_tax", setup_handlers.CreateExpandedWithholdingTax)
					setupApi.Put("/expanded_tax", setup_handlers.UpdateExpandedWithholdingTax)
					setupApi.Delete("/expanded_tax", setup_handlers.DeleteExpandedWithholdingTax)
				}
				// =========================== END =============================

				setupApi.Get("/boq/qq", setup_handlers.GetQQnotes)
				setupApi.Post("/boq", setup_handlers.CreateItemBoq)
				setupApi.Put("/boq", setup_handlers.UpdateItemBoq)

				setupApi.Get("/wiringnotes", setup_handlers.GetWiringNotes)
				setupApi.Post("/wiringnotes", setup_handlers.CreateWiringNote)
				setupApi.Put("/wiringnotes", setup_handlers.UpdateWiringNote)
			}

			// Sales Endpoints
			salesApi := api.Group("/sales")
			{
				salesApi.Get("/quotation", sales_handlers.GetSalesQuotations)
				salesApi.Get("/quotation/latest", sales_handlers.GetLatestQuotations)
				salesApi.Get("/quotation/customers", sales_handlers.GetBpiCustomers)
				//salesApi.Get("/quotation/:id", sales_handlers.GetSalesQuotation)
				//salesApi.Get("/quotation/:id", sales_handlers.GetBpi)
				//salesApi.Post("child/quotation", sales_handlers.CreateSalesQuotationChild)
				// POST for Parent
				salesApi.Post("/quotation", sales_handlers.CreateSalesQuotation)
				salesApi.Put("/quotation", sales_handlers.UpdateQuotation)

				salesApi.Post("/salescanvas", sales_handlers.CreateSalesCanvasSheet)
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
				//salesApi.Post("/project_item", sales_handlers.CreateNewProjectItem)

				// project update
				salesApi.Put("/projects", sales_handlers.UpdateSalesProject)
				salesApi.Put("/project_conditions", sales_handlers.UpdateProjectCondition)
				salesApi.Put("/project_contents", sales_handlers.UpdateProjectContent)

				// salesApi.Post("/project_items", sales_handlers.CreateProjectItem)
				// salesApi.Put("/project_items", sales_handlers.UpdateProjectItem)

				// salesApi.Post("/project_wiring", sales_handlers.CreateProjectWirings)
				// salesApi.Put("/project_wiring", sales_handlers.UpdateProjectWiring)

				salesApi.Get("/projects_pumps", sales_handlers.GetItemPumps)

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

				// Temporary  Completely S.O to DR Endpoints

				salesApi.Get("/order_dr/:id", sales_handlers.GetSalesOrderDR)
				salesApi.Get("/order_dr", sales_handlers.GetSalesOrdersDr)
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
				purchasingApi.Get("/so_purchase_list", purchasing_handlers.GetSOPurchasingList)
				purchasingApi.Get("/pr_purchase_list", purchasing_handlers.GetPRPurchasingList)
				purchasingApi.Get("/purchase_list_supplier", purchasing_handlers.GetSOPurchasingListSupplier)
				purchasingApi.Get("/purchase_guiding_price", purchasing_handlers.GetPurchasingGuidingPrice)
				purchasingApi.Get("/purchase_active_po", purchasing_handlers.GetPurchasingActivePO)
				purchasingApi.Get("/purchase_closed_po", purchasing_handlers.GetPurchasingClosedPO)

				//Purchasing SO Canvass Sheet
				purchasingApi.Get("/purchase_canvass_sheet_so", purchasing_handlers.GetPurchasingCanvassSheetSO)
				purchasingApi.Post("/purchase_canvass_sheet_so", purchasing_handlers.CreatePurchasingCanvassSheet)
				purchasingApi.Put("/purchase_canvass_sheet_so", purchasing_handlers.UpdatePurchasingCanvassSheet)
				purchasingApi.Delete("/purchase_canvass_sheet_so", purchasing_handlers.DeletePurchasingCanvassSheetSupplier)

				purchasingApi.Get("/purchase_order", purchasing_handlers.GetPurchaseOrder)
				purchasingApi.Post("/purchase_order", purchasing_handlers.CreatePurchaseOrder)
				purchasingApi.Put("/purchase_order", purchasing_handlers.UpdatePurchaseOrder)
			}

			//ROUTES
			routes.SetupRoutes(app)

			//Bpi Endpoints

			api.Get("/bpi/entity", bpi_handlers.GetBpiEntityRecords)
			api.Get("/bpi/list", bpi_handlers.GetBpiItemList)
			api.Post("/bpi", bpi_handlers.CreateBpi)
			api.Post("/bpi/createbpi", bpi_handlers.CreateBpiParentFromBranch)
			api.Put("/bpi", bpi_handlers.UpdateBpi)
			api.Put("/bpi/main", bpi_handlers.UpdateBpiMainBranch)
			api.Get("/bpi/:id", sales_handlers.GetBpi)
			api.Get("/bpi_suppliers", sales_handlers.GetBpiSuppliers) // change naming convention to snake case

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

					testApi := setupApi.Group("/test")
					{
						testApi.Get("", websocket.New(func(c *websocket.Conn) {
							services.HandleProjectWs(c, sales_handlers.WsProjects)
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

				// Engineering Endpoints
				engineeringApi := ws.Group("/engineering")
				{
					redboxlistApi := engineeringApi.Group("/redboxlist")
					{
						redboxlistApi.Get("/quotation", websocket.New(func(c *websocket.Conn) {
							services.HandleWsQuotation(c, engineering_handlers.WsgetRedboxQuotationList)
						}))

						redboxlistApi.Get("/job_order/:userId", websocket.New(func(c *websocket.Conn) {
							userId := c.Params("userId")
							services.HandleWsJobOrder(c, func(conn *websocket.Conn) {
								engineering_handlers.WsgetRedboxJobOrder(conn, userId)
							})
						}))
					}
				}
			}
		}
	}

	// Initialize a new Hub instance
	//h := hub.NewHub()

	// Start the hub in its own goroutine
	//go h.Run()

	// Initialize a new Hub instance
	//h := hub.NewHub()

	// Start the hub in its own goroutine
	//go h.Run()

	return app
	// Start Listen
	//app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
