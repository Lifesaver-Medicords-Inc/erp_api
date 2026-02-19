package dispatching_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type CalendarCostTypeHandler struct {
	CalendarCostTypeService *dispatching_services.CalendarCostTypeService
}

func NewCalendarCostTypeHandler(service *dispatching_services.CalendarCostTypeService) *CalendarCostTypeHandler {
	return &CalendarCostTypeHandler{
		CalendarCostTypeService: service,
	}
}

// ✅ GET /calendar-cost-types
func (h *CalendarCostTypeHandler) GetCalendarCostTypesHandler(c *fiber.Ctx) error {
	conditions := map[string]interface{}{}

	costTypes, status, err :=
		h.CalendarCostTypeService.GetCalendarCostTypesService(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, costTypes)
}

// ✅ GET /calendar-cost-types/:id
func (h *CalendarCostTypeHandler) GetCalendarCostTypeHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	schedule, status, err := h.CalendarCostTypeService.GetCalendarCostTypeService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, schedule)
}

// ✅ POST /calendar-cost-types
func (h *CalendarCostTypeHandler) CreateCalendarCostTypeService(c *fiber.Ctx) error {
	var body dispatching_models.CalendarCostTypeModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.CalendarCostTypeService.CreateCalendarCostTypeService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ PUT /calendar-cost-types/:id
func (h *CalendarCostTypeHandler) UpdateCalendarCostTypeHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body dispatching_models.CalendarCostTypeModel
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

	data, status, err := h.CalendarCostTypeService.UpdateCalendarCostTypeService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ DELETE /calendar-cost-types/:id
func (h *CalendarCostTypeHandler) DeleteCalendarCostTypeHandler(c *fiber.Ctx) error {
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

	data, status, err := h.CalendarCostTypeService.DeleteCalendarCostTypeService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
