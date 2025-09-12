package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func PermissionRoutes(app *fiber.App) {
	api := app.Group("/api/permissions")

	permissionService := adminservices.NewPermissionService()
	permissionHandler := adminhandlers.NewPermissionHandler(permissionService)
	api.Get("/", permissionHandler.GetPermissionsHandler)
	api.Get("/:id", permissionHandler.GetPermissionHandler)
	api.Get("/:id", permissionHandler.GetUserPermissionHandler)
	api.Post("/", permissionHandler.CreatePermissionHandler)
	api.Put("/:id", permissionHandler.UpdatePermissionHandler)
	api.Delete("/:id", permissionHandler.DeletePermissionHandler)

}
