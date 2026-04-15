package receiving_report_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services/receiving_report_services"
	"github.com/pierceperado/smpc/utils"
)

type ReceivingReportHandler struct {
	Service *receiving_report_services.ReceivingReportService
}

func NewReceivingReportHandler(service *receiving_report_services.ReceivingReportService) *ReceivingReportHandler {
	return &ReceivingReportHandler{Service: service}
}

func (h *ReceivingReportHandler) GetReceivingReport(c *fiber.Ctx) error {
	data, status, err := h.Service.GetReceivingReport(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) GetWarehouseReceiving(c *fiber.Ctx) error {
	data, status, err := h.Service.GetWarehouseReceiving(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) GetWarehouseAreaReceiving(c *fiber.Ctx) error {
	idParam := c.Params("warehouse_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"WarehouseId": idNum,
	}

	data, status, err := h.Service.GetWarehouseAreaReceiving(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) GetReceivingPODoc(c *fiber.Ctx) error {
	data, status, err := h.Service.GetReceivingPODoc(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) GetReceivingPO(c *fiber.Ctx) error {
	idParam := c.Params("purchase_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"PurchaseId": idNum,
	}

	data, status, err := h.Service.GetReceivingPO(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) CreateReceivingReport(c *fiber.Ctx) error {
	var body inventory_models.ReceivingReportBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateReceivingReport(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) UpdateReceivingReport(c *fiber.Ctx) error {
	var body inventory_models.ReceivingReportBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateReceivingReport(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ReceivingReportHandler) DeleteReceivingReport(c *fiber.Ctx) error {
	var body inventory_models.ReceivingReportBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteReceivingReport(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
