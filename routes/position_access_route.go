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
	api.Get("/", positionAccessHandler.GetAllPositionAccess)
	api.Get("/:id", positionAccessHandler.GetPositionAccess)
	api.Post("", positionAccessHandler.CreatePositionAccess)
	api.Put("/:id", positionAccessHandler.UpdatePositionAccess)
	api.Delete("/:id", positionAccessHandler.DeletePositionAccess)
	api.Post("/update-all-access/:id", positionAccessHandler.UpdatePositionAllAccess)

}
