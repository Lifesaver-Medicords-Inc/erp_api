package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/bin_location_handlers"
	"github.com/pierceperado/smpc/handlers/engineering_handlers"
	item_request_handlers "github.com/pierceperado/smpc/handlers/item_request_handlers2"
	"github.com/pierceperado/smpc/handlers/job_order_handlers"
	pick_activity_handlers "github.com/pierceperado/smpc/handlers/pick_activity_handlers2"
	"github.com/pierceperado/smpc/services/bin_location_services"
	item_request_services "github.com/pierceperado/smpc/services/item_request_services2"
	"github.com/pierceperado/smpc/services/job_order_services"
	pick_activity_services "github.com/pierceperado/smpc/services/pick_activity_services2"
)

func EngineeringRoutes(router fiber.Router) {
	engineeringApi := router.Group("/engineering")

	setupJobOrderRoutes(engineeringApi)
	setupItemRequestRoutes(engineeringApi)
	setupBinLocationRoutes(engineeringApi)
	setupPickActivityRoutes(engineeringApi)
	setupSalesQuotationListRoutes(engineeringApi)
}

// §3.2 Sales Quotation List (Phase 4 item 4.1, "build the list" half) - the
// engineer's own scoped view of REQUEST FOR ENGR. quotations.
func setupSalesQuotationListRoutes(api fiber.Router) {
	api.Get("/sales_quotation_list/:userId", engineering_handlers.GetEngineeringQuotationListByEngr)
}

func setupJobOrderRoutes(api fiber.Router) {
	handler := job_order_handlers.NewJobOrderHandler(job_order_services.NewJobOrderService())
	api.Get("/job_order/engr_list", handler.GetEngineerList)
	api.Get("/job_order/sales_order", handler.GetSalesOrderViewEng)
	api.Get("/job_order/components/:bom_id", handler.GetComponents)
	api.Post("/job_order", handler.CreateJobOrder)
	api.Put("/job_order", handler.UpdateJobOrder)
	api.Post("/job_order/:id/accept", handler.AcceptJobOrder)
	api.Post("/job_order/:id/acknowledge", handler.AcknowledgeJobOrder)
	// Must be registered before the /job_order/:user_id wildcard below, or Fiber
	// would try to parse "pending_production_reports" as a user id.
	api.Get("/job_order/pending_production_reports", handler.GetPendingProductionReports)
	api.Get("/job_order/:user_id", handler.GetJobOrder)
}

func setupItemRequestRoutes(api fiber.Router) {
	handler := item_request_handlers.NewItemRequestHandler(item_request_services.NewItemRequestService())
	api.Get("/item_request", handler.GetItemRequest)
	api.Get("/item_request/items", handler.GetAllItems)
	api.Get("/item_request/users", handler.GetUserList)
	api.Get("/item_request/sales_order_doc", handler.GetItemReqSODoc)
	api.Get("/item_request/sales_order/:sales_id", handler.GetItemReqSO)
	api.Post("/item_request", handler.CreateItemRequest)
	api.Put("/item_request", handler.UpdateItemRequest)
	api.Delete("/item_request", handler.DeleteItemRequest)
}

func setupPickActivityRoutes(api fiber.Router) {
	handler := pick_activity_handlers.NewPickActivityHandler(pick_activity_services.NewPickActivityService())
	api.Get("/pick_activity", handler.GetPickActivity)
	api.Get("/pick_activity/warehouse", handler.GetWarehousePickAct)
	api.Get("/pick_activity/warehouse_area/:warehouse_id", handler.GetWarehouseAreaPickAct)
	api.Get("/pick_activity/sales_order_doc", handler.GetPickActSODoc)
	api.Get("/pick_activity/sales_order/:sales_id", handler.GetPickActSO)
	api.Post("/pick_activity", handler.CreatePickActivity)
	api.Put("/pick_activity", handler.UpdatePickActivity)
	api.Delete("/pick_activity", handler.DeletePickActivity)
}

func setupBinLocationRoutes(api fiber.Router) {
	handler := bin_location_handlers.NewBinLocationHandler(bin_location_services.NewBinLocationService())
	api.Get("/bin_location/:item_id", handler.GetBinLocations)
}
