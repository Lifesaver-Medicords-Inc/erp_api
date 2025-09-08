package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func PositionRoutes(app *fiber.App) {
	api := app.Group("/api/positions")

	positionService := adminservices.NewPositionService()
	positionHandler := adminhandlers.NewPositionHandler(positionService)
	api.Get("/", positionHandler.GetPositions)
	api.Get("/:id", positionHandler.GetPosition)
	api.Post("/", positionHandler.CreatePosition)
	api.Put("/:id", positionHandler.UpdatePosition)
	api.Delete("/:id", positionHandler.DeletePosition)

}
