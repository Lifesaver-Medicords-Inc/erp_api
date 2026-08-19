package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func AccessModuleRoutes(app *fiber.App) {
	api := app.Group("/api/access-modules")

	accessModuleService := adminservices.NewAccessModuleService()
	accessModuleHandler := adminhandlers.NewAccessModuleHandler(accessModuleService)
	api.Get("/", accessModuleHandler.GetAllAccessModulesHandler)
}
