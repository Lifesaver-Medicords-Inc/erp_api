package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func ItemReleaseRoutes(app *fiber.App) {
	releases := app.Group("/api/item-releases")

	itemReleasService := dispatching_services.NewItemReleaseService()
	itemReleaseHandler := dispatching_handlers.NewItemReleaseHandler(itemReleasService)
	releases.Get("/sales-order-details/", itemReleaseHandler.GetSalesOrderItemReleaseDetailsHandler)
	releases.Get("/", itemReleaseHandler.GetItemReleasesHandler)
	releases.Get("/:id", itemReleaseHandler.GetItemReleaseHandler)
	releases.Post("/", itemReleaseHandler.CreateItemReleaseHandler)
	releases.Put("/:id", itemReleaseHandler.UpdateItemReleaseHandler)
	releases.Delete("/:id", itemReleaseHandler.DeleteItemReleaseHandler)
}
