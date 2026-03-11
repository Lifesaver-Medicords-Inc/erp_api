package dispatching_handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
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
	var body dispatching_models.DeliveryReceipt


	if err := c.BodyParser(&body); err != nil {
		fmt.Println("Error parsing body:", err)
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.Service.CreateDeliveryReceiptService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	fmt.Println("DR data", data)

	return utils.RespondSuccess(c, data)
}

func (h *DeliveryReceiptHandler) UpdateDeliveryReceiptHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	var body dispatching_models.DeliveryReceipt
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	conditions := map[string]interface{}{
		"id": id,
	}
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	data, status, err := h.Service.UpdateDeliveryReceiptService(&body, conditions, at)
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

	conditions := map[string]interface{}{
		"id": id,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.Service.DeleteDeliveryReceiptService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *DeliveryReceiptHandler) GetSOWithApprovedIRHandler(c *fiber.Ctx) error {

	data, code, err := h.Service.GetSOWithApprovedIRService(nil)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func (h *DeliveryReceiptHandler) GetSOWithApprovedIRDetailsHandler(c *fiber.Ctx) error {

	idParam := c.Params("item_release_id")

	itemReleaseID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid item_release_id")
	}

	data, status, err := h.Service.GetSOWithApprovedIRDetailsService(itemReleaseID)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
