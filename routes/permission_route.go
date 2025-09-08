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
	api.Get("/", permissionHandler.GetPermissions)
	api.Get("/:id", permissionHandler.GetPermission)
	api.Get("/:id", permissionHandler.GetUserPermission)
	api.Post("/", permissionHandler.CreatePermission)
	api.Put("/:id", permissionHandler.UpdatePermission)
	api.Delete("/:id", permissionHandler.DeletePermission)

}
