package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/middlewares"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.MigrateDb()
}

func main() {
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	api := app.Group("/api")
	{
		// Public Routes
		api.Post("/register", handlers.Register)
		api.Post("/login", handlers.Login)
		api.Post("/logout", handlers.Logout)

		// Protected Routes
		api.Use(middlewares.RequireAuth)
		{
			api.Get("/profile", handlers.GetUserProfile)
		}
	}

	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
