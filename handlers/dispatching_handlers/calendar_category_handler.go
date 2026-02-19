package dispatching_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type CalendarCategoryHandler struct {
	CalendarCategoryService *dispatching_services.CalendarCategoryService
}

func NewCalendarCategoryHandler(service *dispatching_services.CalendarCategoryService) *CalendarCategoryHandler {
	return &CalendarCategoryHandler{
		CalendarCategoryService: service,
	}
}

// ✅ GET /calendar-categories
func (h *CalendarCategoryHandler) GetCalendarCategoriesHandler(c *fiber.Ctx) error {
	conditions := map[string]interface{}{}

	categories, status, err :=
		h.CalendarCategoryService.GetCalendarCategoriesService(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, categories)
}

// ✅ GET /calendar-categories/:id
func (h *CalendarCategoryHandler) GetCalendarCategoryHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	category, status, err := h.CalendarCategoryService.GetCalendarCategoryService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, category)
}

// ✅ POST /calendar-categories
func (h *CalendarCategoryHandler) CreateCalendarCategoryHandler(c *fiber.Ctx) error {
	var body dispatching_models.CalendarCategoryModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.CalendarCategoryService.CreateCalendarCategoryService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ PUT /calendar-categories/:id
func (h *CalendarCategoryHandler) UpdateCalendarCategoryHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body dispatching_models.CalendarCategoryModel
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

	data, status, err := h.CalendarCategoryService.UpdateCalendarCategoryService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ✅ DELETE /calendar-categories/:id
func (h *CalendarCategoryHandler) DeleteCalendarCategoryHandler(c *fiber.Ctx) error {
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

	data, status, err := h.CalendarCategoryService.DeleteCalendarCategoryService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
