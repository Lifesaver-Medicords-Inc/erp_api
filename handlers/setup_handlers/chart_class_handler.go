package setup_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

type ChartClassHandler struct {
	Service *setup_services.ChartClassService
}

func NewChartClassHandler(service *setup_services.ChartClassService) *ChartClassHandler {
	return &ChartClassHandler{Service: service}
}

func (h *ChartClassHandler) GetChartClasses(c *fiber.Ctx) error {
	search := c.Query("search")

	var id int
	if idParam := c.Query("id"); idParam != "" {
		var err error
		id, err = strconv.Atoi(idParam)
		if err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "invalid id")
		}
	}

	data, status, pagination, err := h.Service.GetChartClasses(nil, search, id)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
}

func (h *ChartClassHandler) GetChartClass(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := h.Service.GetChartClass(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ChartClassHandler) CreateChartClass(c *fiber.Ctx) error {
	var body accounting_models.ChartClass

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateChartClass(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ChartClassHandler) UpdateChartClass(c *fiber.Ctx) error {
	var body accounting_models.ChartClass

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateChartClass(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ChartClassHandler) DeleteChartClass(c *fiber.Ctx) error {
	var body accounting_models.ChartClass

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteChartClass(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
