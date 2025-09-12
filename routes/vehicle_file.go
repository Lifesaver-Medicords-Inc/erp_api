package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func VehicleFileRoutes(app *fiber.App) {
	api := app.Group("/api/vehicle-files")

	uploadDir := "./files"
	fileService := adminservices.NewVehicleFileService(uploadDir)
	fileHandler := adminhandlers.NewVehicleFileHandler(fileService)
	api.Get("/download", fileHandler.DownloadFileHandler)
	api.Post("/", fileHandler.UploadVehicleFileHandler)
	api.Get("/", fileHandler.GetVehicleFilesHandler)
	api.Get("/:id", fileHandler.GetVehicleFileHandler)
	api.Delete("/:id", fileHandler.DeleteVehicleFileHandler)
}
