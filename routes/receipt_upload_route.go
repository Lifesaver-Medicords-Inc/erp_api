package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
)

func ReceiptUploadRoutes(app *fiber.App) {
	api := app.Group("/api/uploads")
	api.Post("/receipts", dispatching_handlers.UploadReceiptHandler)
}
