package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func PositionAccessRoutes(app *fiber.App) {
	api := app.Group("/api/position-access")

	positionAccessService := adminservices.NewPositionAccessService()
	positionAccessHandler := adminhandlers.NewPositionAccessHandler(positionAccessService)
	api.Get("/", positionAccessHandler.GetAllPositionAccessHandler)
	api.Get("/:id", positionAccessHandler.GetPositionAccessHandler)
	api.Post("", positionAccessHandler.CreatePositionAccessHandler)
	api.Put("/:id", positionAccessHandler.UpdatePositionAccessHandler)
	api.Delete("/:id", positionAccessHandler.DeletePositionAccessHandler)
	api.Post("/update-all-access/:id", positionAccessHandler.UpdatePositionAllAccessHandler)
}
