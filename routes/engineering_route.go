package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/bin_location_handlers"
	item_request_handlers "github.com/pierceperado/smpc/handlers/item_request_handlers2"
	"github.com/pierceperado/smpc/handlers/job_order_handlers"
	"github.com/pierceperado/smpc/services/bin_location_services"
	item_request_services "github.com/pierceperado/smpc/services/item_request_services2"
	"github.com/pierceperado/smpc/services/job_order_services"
)

func EngineeringRoutes(router fiber.Router) {
	engineeringApi := router.Group("/engineering")

	setupJobOrderRoutes(engineeringApi)
	setupItemRequestRoutes(engineeringApi)
	setupBinLocationRoutes(engineeringApi)
}

func setupJobOrderRoutes(api fiber.Router) {
	handler := job_order_handlers.NewJobOrderHandler(job_order_services.NewJobOrderService())
	api.Get("/job_order/engr_list", handler.GetEngineerList)
	api.Get("/job_order/sales_order", handler.GetSalesOrderViewEng)
	api.Get("/job_order/components/:bom_id", handler.GetComponents)
	api.Post("/job_order", handler.CreateJobOrder)
	api.Put("/job_order", handler.UpdateJobOrder)
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

func setupBinLocationRoutes(api fiber.Router) {
	handler := bin_location_handlers.NewBinLocationHandler(bin_location_services.NewBinLocationService())
	api.Get("/bin_location/:item_id", handler.GetBinLocations)
}
