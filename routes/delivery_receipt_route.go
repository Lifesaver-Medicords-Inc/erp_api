package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func DeliveryReceiptRoutes(app *fiber.App) {
	api := app.Group("/api/delivery-receipts")

	calendarService := dispatching_services.NewCalendarScheduleService()
	deliveryReceiptService := dispatching_services.NewDeliveryReceiptService(calendarService)
	deliveryReceiptHandler := dispatching_handlers.NewDeliveryReceiptHandler(deliveryReceiptService)
	api.Get("/", deliveryReceiptHandler.GetDeliveryReceiptsHandler)
	api.Get("/so-with-approved-ir/", deliveryReceiptHandler.GetSOWithApprovedIRHandler)
	api.Get("/so-with-approved-ir-details/:item_release_id", deliveryReceiptHandler.GetSOWithApprovedIRDetailsHandler)
	api.Get("/:id", deliveryReceiptHandler.GetDeliveryReceiptHandler)
	api.Post("/", deliveryReceiptHandler.CreateDeliveryReceiptHandler)
	api.Put("/:id", deliveryReceiptHandler.UpdateDeliveryReceiptHandler)
	api.Delete("/:id", deliveryReceiptHandler.DeleteDeliveryReceiptHandler)
}
