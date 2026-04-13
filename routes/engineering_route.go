package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/job_order_handlers"
	"github.com/pierceperado/smpc/services/job_order_services"
)

func EngineeringRoutes(router fiber.Router) {
	engineeringApi := router.Group("/engineering")

	setupJobOrderRoutes(engineeringApi)
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
