package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/purchase_return_handlers"
	"github.com/pierceperado/smpc/services/purchase_return_services"
)

// PurchaseReturnRoutes - Purchase Return (PRT#), spec §5.8. Own top-level
// group, matching SalesReturnRoutes' convention. No approve endpoint yet -
// see PurchaseReturnService.CreatePurchaseReturn's doc comment for why.
func PurchaseReturnRoutes(app *fiber.App) {
	api := app.Group("/api/purchase-returns")

	handler := purchase_return_handlers.NewPurchaseReturnHandler(purchase_return_services.NewPurchaseReturnService())

	api.Get("/", handler.GetPurchaseReturn)
	api.Get("/:id", handler.GetPurchaseReturnById)
	api.Post("/", handler.CreatePurchaseReturn)
}
