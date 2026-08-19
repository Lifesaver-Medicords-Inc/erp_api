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
	// GetUserPermissionHandler (looks up by user_id, not the permission row's own id) was
	// registered on this exact same "/:id" path above - Fiber only ever dispatches to the
	// first match, so it was dead code and every "get this user's permission" caller
	// (e.g. LoginForm.cs on login) was actually hitting GetPermissionHandler instead,
	// which filters by the wrong column. Given its own path so both are reachable.
	api.Get("/user/:id", permissionHandler.GetUserPermissionHandler)
	api.Post("/", permissionHandler.CreatePermissionHandler)
	api.Put("/:id", permissionHandler.UpdatePermissionHandler)
	api.Delete("/:id", permissionHandler.DeletePermissionHandler)
}
