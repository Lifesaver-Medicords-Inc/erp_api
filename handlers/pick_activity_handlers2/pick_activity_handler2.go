package pick_activity_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	pick_activity_services "github.com/pierceperado/smpc/services/pick_activity_services2"
	"github.com/pierceperado/smpc/utils"
)

type PickActivityHandler struct {
	Service *pick_activity_services.PickActivityService
}

func NewPickActivityHandler(service *pick_activity_services.PickActivityService) *PickActivityHandler {
	return &PickActivityHandler{Service: service}
}

func (h *PickActivityHandler) GetPickActivity(c *fiber.Ctx) error {

	data, status, err := h.Service.GetPickActivity(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) GetWarehousePickAct(c *fiber.Ctx) error {
	data, status, err := h.Service.GetWarehousePickAct(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) GetWarehouseAreaPickAct(c *fiber.Ctx) error {
	idParam := c.Params("warehouse_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"WarehouseId": idNum,
	}

	data, status, err := h.Service.GetWarehouseAreaPickAct(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) GetPickActSODoc(c *fiber.Ctx) error {
	data, status, err := h.Service.GetPickActSODoc(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) GetPickActSO(c *fiber.Ctx) error {
	idParam := c.Params("sales_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"SalesId": idNum,
	}

	data, status, err := h.Service.GetPickActSO(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) CreatePickActivity(c *fiber.Ctx) error {
	var body inventory_models.PickActivityBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreatePickActivity(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) UpdatePickActivity(c *fiber.Ctx) error {
	var body inventory_models.PickActivityBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdatePickActivity(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PickActivityHandler) DeletePickActivity(c *fiber.Ctx) error {
	var body inventory_models.PickActivityBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeletePickActivity(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
