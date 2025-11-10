package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func CalendarScheduleRoutes(app *fiber.App) {
	api := app.Group("/api/calendar-schedules")

	calendarScheduleService := dispatching_services.NewCalendarScheduleService()
	calendarSchedule := dispatching_handlers.NewCalendarScheduleHandler(calendarScheduleService)
	api.Get("/", calendarSchedule.GetCalendarSchedulesHandler)
	api.Get("/:id", calendarSchedule.GetCalendarScheduleHandler)
	api.Post("/", calendarSchedule.CreateCalendarScheduleHandler)
	api.Put("/:id", calendarSchedule.UpdateCalendarScheduleHandler)
	api.Delete("/:id", calendarSchedule.DeleteCalendarScheduleHandler)

}
