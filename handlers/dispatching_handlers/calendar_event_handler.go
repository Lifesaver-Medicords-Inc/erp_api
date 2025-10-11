package dispatching_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type CalendarEventHandler struct {
	CalendarEventService *dispatching_services.CalendarEventService
}

func NewCalendarEventHandler(service *dispatching_services.CalendarEventService) *CalendarEventHandler {
	return &CalendarEventHandler{
		CalendarEventService: service,
	}
}

// ✅ GET /calendar-events
func (h *CalendarEventHandler) GetCalendarEventsHandler(c *fiber.Ctx) error {
	id := c.Query("id")
	department := c.Query("department")
	isCancelled := c.Query("is_cancelled")

	conditions := make(map[string]interface{})

	if id != "" {
		idNum, _ := strconv.Atoi(id)
		if idNum != 0 {
			conditions["id"] = idNum
		}
	}

	if department != "" {
		conditions["department"] = department
	}

	if isCancelled != "" {
		if isCancelled == "true" {
			conditions["is_cancelled"] = true
		} else {
			conditions["is_cancelled"] = false
		}
	}

	events, status, err := h.CalendarEventService.GetCalendarEventsService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, events)
}

// ✅ GET /calendar-events/:id
func (h *CalendarEventHandler) GetCalendarEventHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	event, status, err := h.CalendarEventService.GetCalendarEventService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, event)
}

// ✅ POST /calendar-events
func (h *CalendarEventHandler) CreateCalendarEventHandler(c *fiber.Ctx) error {
	var body models.CalendarEventModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.CalendarEventService.CreateCalendarEventService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ PUT /calendar-events/:id
func (h *CalendarEventHandler) UpdateCalendarEventHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body models.CalendarEventModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	data, status, err := h.CalendarEventService.UpdateCalendarEventService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ DELETE /calendar-events/:id
func (h *CalendarEventHandler) DeleteCalendarEventHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.CalendarEventService.DeleteCalendarEventService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
