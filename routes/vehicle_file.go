package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func FileRoutes(app *fiber.App) {
	api := app.Group("/api/vehicles-files")

	uploadDir := "./files"
	fileService := adminservices.NewVehicleFileService(uploadDir)
	fileHandler := adminhandlers.NewVehicleFileHandler(fileService)
	api.Post("/", fileHandler.UploadHandler)
}
