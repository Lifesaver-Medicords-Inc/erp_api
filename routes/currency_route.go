package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func CurrencyRoutes(app *fiber.App) {
	api := app.Group("/api/currency")

	currencyService := adminservices.NewCurrencyService()
	currencyHandler := adminhandlers.NewCurrencyHandler(currencyService)

	api.Get("/:id", currencyHandler.GetCurrencyHandler)
}
