package routes

import (
	"github.com/gofiber/fiber/v2"
	dispatching_handlers "github.com/pierceperado/smpc/handlers/dispatching_handlers"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
)

func CalendarEventRoutes(app *fiber.App) {
	api := app.Group("/calendar-event")

	calendarEventService := dispatching_services.NewCalendarEventService()
	calendarEvent := dispatching_handlers.NewCalendarEventHandler(calendarEventService)
	api.Get("/", calendarEvent.GetCalendarEventsHandler)
	api.Get("/:id", calendarEvent.GetCalendarEventHandler)
	api.Post("/", calendarEvent.CreateCalendarEventHandler)
	api.Put("/:id", calendarEvent.UpdateCalendarEventHandler)
	api.Delete("/:id", calendarEvent.DeleteCalendarEventHandler)

}
