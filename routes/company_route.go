package routes

import (
	"github.com/gofiber/fiber/v2"
	adminhandlers "github.com/pierceperado/smpc/handlers/admin_handlers"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

func CompanyRoutes(app *fiber.App) {
	api := app.Group("/api/companies")

	companyService := adminservices.NewCompanyService()
	companyHandler := adminhandlers.NewCompanyHandler(companyService)

	api.Get("/", companyHandler.GetCompaniesHandler)
	api.Get("/:id", companyHandler.GetCompanyHandler)
	api.Post("/", companyHandler.CreateCompanyHandler)
	api.Put("/:id", companyHandler.UpdateCompanyHandler)
	api.Delete("/:id", companyHandler.DeleteCompanyHandler)
}
