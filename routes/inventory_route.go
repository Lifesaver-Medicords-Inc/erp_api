package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/receiving_report_handlers"
	"github.com/pierceperado/smpc/services/receiving_report_services"
)

func InventoryRoutes(router fiber.Router) {
	inventoryApi := router.Group("/inventory")

	setupReceivingReportRoutes(inventoryApi)
}

func setupReceivingReportRoutes(api fiber.Router) {
	handler := receiving_report_handlers.NewReceivingReportHandler(receiving_report_services.NewReceivingReportService())
	api.Get("/receiving_report", handler.GetReceivingReport)
	api.Get("/receiving_report/warehouse", handler.GetWarehouseReceiving)
	api.Get("/receiving_report/warehouse_area/:warehouse_id", handler.GetWarehouseAreaReceiving)
	api.Get("/receiving_report/purchase_order_doc", handler.GetReceivingPODoc)
	api.Get("/receiving_report/purchase_order/:purchase_id", handler.GetReceivingPO)
	api.Post("/receiving_report", handler.CreateReceivingReport)
	api.Put("/receiving_report", handler.UpdateReceivingReport)
	api.Delete("/receiving_report", handler.DeleteReceivingReport)
}
