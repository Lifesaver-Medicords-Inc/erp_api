package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func CalendarScheduleRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Calendar Schedules
	calendarScheduleService := dispatching_services.NewCalendarScheduleService()
	calendarScheduleHandler := dispatching_handlers.NewCalendarScheduleHandler(calendarScheduleService)

	calendar := api.Group("/calendar-schedules")
	calendar.Get("/", calendarScheduleHandler.GetCalendarSchedulesHandler)
	calendar.Get("/:id", calendarScheduleHandler.GetCalendarScheduleHandler)
	calendar.Post("/", calendarScheduleHandler.CreateCalendarScheduleHandler)
	calendar.Put("/:id", calendarScheduleHandler.UpdateCalendarScheduleHandler)
	calendar.Delete("/sales/:id", calendarScheduleHandler.DeleteSalesScheduleHandler)
	calendar.Delete("/engineering/:id", calendarScheduleHandler.DeleteEngineeringScheduleHandler)
	calendar.Delete("/logistics/:id", calendarScheduleHandler.DeleteLogisticsScheduleHandler)

	// Calendar Categories / Setups
	calendarCategoryService := dispatching_services.NewCalendarCategoryService()
	calendarCategoryHandler := dispatching_handlers.NewCalendarCategoryHandler(calendarCategoryService)

	categories := api.Group("/calendar-categories")
	categories.Get("/", calendarCategoryHandler.GetCalendarCategoriesHandler)
	categories.Get("/:id", calendarCategoryHandler.GetCalendarCategoryHandler)
	categories.Post("/", calendarCategoryHandler.CreateCalendarCategoryHandler)
	categories.Put("/:id", calendarCategoryHandler.UpdateCalendarCategoryHandler)
	categories.Delete("/:id", calendarCategoryHandler.DeleteCalendarCategoryHandler)

	// Calendar Cost Types / Setups
	calendarCostTypeService := dispatching_services.NewCalendarCostTypeService()
	costTypesHandler := dispatching_handlers.NewCalendarCostTypeHandler(calendarCostTypeService)

	costTypes := api.Group("/calendar-cost-types")
	costTypes.Get("/", costTypesHandler.GetCalendarCostTypesHandler)
	costTypes.Get("/:id", costTypesHandler.GetCalendarCostTypeHandler)
	costTypes.Post("/", costTypesHandler.CreateCalendarCostTypeService)
	costTypes.Put("/:id", costTypesHandler.UpdateCalendarCostTypeHandler)
	costTypes.Delete("/:id", costTypesHandler.DeleteCalendarCostTypeHandler)
}
