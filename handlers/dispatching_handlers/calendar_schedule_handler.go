package dispatching_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type CalendarScheduleHandler struct {
	CalendarScheduleService *dispatching_services.CalendarScheduleService
}

func NewCalendarScheduleHandler(service *dispatching_services.CalendarScheduleService) *CalendarScheduleHandler {
	return &CalendarScheduleHandler{
		CalendarScheduleService: service,
	}
}

// ✅ GET /calendar-schedules
func (h *CalendarScheduleHandler) GetCalendarSchedulesHandler(c *fiber.Ctx) error {
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

	schedules, status, err := h.CalendarScheduleService.GetCalendarSchedulesService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, schedules)
}

// ✅ GET /calendar-schedules/:id
func (h *CalendarScheduleHandler) GetCalendarScheduleHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	schedule, status, err := h.CalendarScheduleService.GetCalendarScheduleService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, schedule)
}

// ✅ POST /calendar-schedules
func (h *CalendarScheduleHandler) CreateCalendarScheduleHandler(c *fiber.Ctx) error {
	var body models.CalendarScheduleModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.CalendarScheduleService.CreateCalendarScheduleService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ PUT /calendar-schedules/:id
func (h *CalendarScheduleHandler) UpdateCalendarScheduleHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body models.CalendarScheduleModel
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

	data, status, err := h.CalendarScheduleService.UpdateCalendarScheduleService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ DELETE /calendar-schedules/:id
func (h *CalendarScheduleHandler) DeleteCalendarScheduleHandler(c *fiber.Ctx) error {
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

	data, status, err := h.CalendarScheduleService.DeleteCalendarScheduleService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
