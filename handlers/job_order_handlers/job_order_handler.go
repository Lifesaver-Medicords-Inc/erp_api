package job_order_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/job_order_services"
	"github.com/pierceperado/smpc/utils"
)

type JobOrderHandler struct {
	Service *job_order_services.JobOrderService
}

func NewJobOrderHandler(service *job_order_services.JobOrderService) *JobOrderHandler {
	return &JobOrderHandler{Service: service}
}

func (h *JobOrderHandler) GetJobOrder(c *fiber.Ctx) error {
	idParam := c.Params("user_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"UserId": idNum,
	}

	data, status, err := h.Service.GetJobOrder(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) GetEngineerList(c *fiber.Ctx) error {
	data, status, err := h.Service.GetEngineerList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) GetSalesOrderViewEng(c *fiber.Ctx) error {
	data, status, err := h.Service.GetSalesOrderViewEng(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) GetComponents(c *fiber.Ctx) error {
	idParam := c.Params("bom_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"BomId": idNum,
	}
	data, status, err := h.Service.GetComponents(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) CreateJobOrder(c *fiber.Ctx) error {
	var body []models.JobOrder

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateJobOrder(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *JobOrderHandler) UpdateJobOrder(c *fiber.Ctx) error {
	var body []models.JobOrder

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateJobOrder(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
