package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func ItemReleaseRoutes(app *fiber.App) {
	api := app.Group("/api/item-releases")

	itemReleasService := dispatching_services.NewItemReleaseService()
	itemReleaseHandler := dispatching_handlers.NewItemReleaseHandler(itemReleasService)
	api.Get("/", itemReleaseHandler.GetItemReleasesHandler)
	api.Get("/:id", itemReleaseHandler.GetItemReleaseHandler)
	api.Post("/", itemReleaseHandler.CreateItemReleaseHandler)
	api.Put("/:id", itemReleaseHandler.UpdateItemReleaseHandler)
	api.Delete("/:id", itemReleaseHandler.DeleteItemReleaseHandler)

}
