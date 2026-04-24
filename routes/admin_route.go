package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func AdminRoutes(router fiber.Router) {
	adminApi := router.Group("/admin")

	setupRedisRoutes(adminApi)
}

func setupRedisRoutes(api fiber.Router) {
	handler := adminhandlers.NewAdminHandler(adminservices.NewRedisService())
	api.Get("/clear_all", handler.ClearAllCache)
}
