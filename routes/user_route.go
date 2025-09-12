package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func UserRoutes(app *fiber.App) {

	api := app.Group("/api/users")

	userService := adminservices.NewUserService()
	permissionService := adminservices.NewPermissionService()
	userHandler := adminhandlers.NewUserHandler(userService, permissionService)
	api.Post("/", userHandler.CreateUserHandler)
	api.Get("/:id", userHandler.GetUserHandler)
	api.Put("/:id", userHandler.UpdateUserHandler)
	api.Delete("/:id", userHandler.DeleteUserHandler)
	api.Get("/", userHandler.GetAllUsersHandler)
	api.Get("/with-position/:id", userHandler.GetPositionUsersHandler)
	api.Put("/position/:id", userHandler.UpdateUserPositionHandler)

}
