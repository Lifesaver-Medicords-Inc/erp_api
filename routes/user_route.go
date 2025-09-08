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
	api.Post("/", userHandler.CreateUser)
	api.Get("/:id", userHandler.GetUser)
	api.Put("/:id", userHandler.UpdateUser)
	api.Delete("/:id", userHandler.DeleteUser)
	api.Get("/", userHandler.GetAllUsers)
	api.Get("/with-position/:id", userHandler.GetPositionUsers)
	api.Put("/position/:id", userHandler.UpdateUserPosition)

}
