package dispatching_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type DeliveryReceiptHandler struct {
	Service *dispatching_services.DeliveryReceiptService
}

func NewDeliveryReceiptHandler(service *dispatching_services.DeliveryReceiptService) *DeliveryReceiptHandler {
	return &DeliveryReceiptHandler{
		Service: service,
	}
}

func (h *DeliveryReceiptHandler) GetDeliveryReceiptsHandler(c *fiber.Ctx) error {
	id := c.Query("id")
	salesOrderID := c.Query("sales_order_id")

	filters := make(map[string]interface{})

	if id != "" {
		filters["id"], _ = strconv.Atoi(id)
	}
	if salesOrderID != "" {
		filters["sales_order_id"], _ = strconv.Atoi(salesOrderID)
	}

	data, status, err := h.Service.GetDeliveryReceiptsService(filters)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *DeliveryReceiptHandler) GetDeliveryReceiptHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	data, status, err := h.Service.GetDeliveryReceiptService(map[string]interface{}{"id": id})
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func (h *DeliveryReceiptHandler) CreateDeliveryReceiptHandler(c *fiber.Ctx) error {
	var body models.DeliveryReceiptModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	data, status, err := h.Service.CreateDeliveryReceiptService(&body)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *DeliveryReceiptHandler) UpdateDeliveryReceiptHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	var body models.DeliveryReceiptModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	data, status, err := h.Service.UpdateDeliveryReceiptService(uint(id), &body)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *DeliveryReceiptHandler) DeleteDeliveryReceiptHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	ok, status, err := h.Service.DeleteDeliveryReceiptService(uint(id))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, ok)
}
