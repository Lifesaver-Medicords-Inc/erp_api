package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pierceperado/smpc/handlers"
	"github.com/pierceperado/smpc/initializers"
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

	// public routes start
	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
	api.Post("/logout", handlers.Logout)
	// public routes end

	app.Listen(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT"))
}
