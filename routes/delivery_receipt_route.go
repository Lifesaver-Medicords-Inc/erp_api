package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func DeliveryReceiptRoutes(app *fiber.App) {
	api := app.Group("/delivery-receipt")

	deliveryReceiptService := dispatching_services.NewDeliveryReceiptService()
	deliveryReceiptHandler := dispatching_handlers.NewDeliveryReceiptHandler(deliveryReceiptService)
	api.Get("/", deliveryReceiptHandler.GetDeliveryReceiptsHandler)
	api.Get("/:id", deliveryReceiptHandler.GetDeliveryReceiptHandler)
	api.Post("/", deliveryReceiptHandler.CreateDeliveryReceiptHandler)
	api.Put("/:id", deliveryReceiptHandler.UpdateDeliveryReceiptHandler)
	api.Delete("/:id", deliveryReceiptHandler.DeleteDeliveryReceiptHandler)

}
