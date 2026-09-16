package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	"github.com/pierceperado/smpc/middlewares"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func UserRoutes(app *fiber.App) {
	api := app.Group("/api/users")

	userService := adminservices.NewUserService()
	permissionService := adminservices.NewPermissionService()
	userHandler := adminhandlers.NewUserHandler(userService, permissionService)

	// Every change to a user needs the Users screen's grant (ADMIN USERS); reads stay open
	// to any signed-in user. Before this, any logged-in user of any app could create an
	// account, move someone to another position, or delete one.
	canManageUsers := middlewares.RequireAccessCode(middlewares.ManageUsersAccessCode)

	api.Post("/", canManageUsers, userHandler.CreateUserHandler)
	api.Get("/:id", userHandler.GetUserHandler)
	api.Put("/:id/password", canManageUsers, userHandler.ChangePasswordHandler)
	api.Put("/:id", canManageUsers, userHandler.UpdateUserHandler)
	api.Delete("/:id", canManageUsers, userHandler.DeleteUserHandler)
	api.Get("/", userHandler.GetAllUsersHandler)
	api.Get("/with-position/:id", userHandler.GetPositionUsersHandler)
	api.Put("/position/:id", canManageUsers, userHandler.UpdateUserPositionHandler)
}
