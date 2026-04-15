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
	api.Get("/", positionHandler.GetPositionsHandler)
	api.Get("/:id", positionHandler.GetPositionHandler)
	api.Post("/", positionHandler.CreatePositionHandler)
	api.Put("/:id", positionHandler.UpdatePositionHandler)
	api.Delete("/:id", positionHandler.DeletePositionHandler)
}
