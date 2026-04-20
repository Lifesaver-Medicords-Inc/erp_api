package item_request_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	item_request_services "github.com/pierceperado/smpc/services/item_request_services2"
	"github.com/pierceperado/smpc/utils"
)

type ItemRequestHandler struct {
	Service *item_request_services.ItemRequestService
}

func NewItemRequestHandler(service *item_request_services.ItemRequestService) *ItemRequestHandler {
	return &ItemRequestHandler{Service: service}
}

func (h *ItemRequestHandler) GetItemRequest(c *fiber.Ctx) error {

	data, status, err := h.Service.GetItemRequest(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) GetUserList(c *fiber.Ctx) error {
	data, status, err := h.Service.GetUserList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) GetAllItems(c *fiber.Ctx) error {
	data, status, err := h.Service.GetAllItems(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) GetItemReqSODoc(c *fiber.Ctx) error {
	data, status, err := h.Service.GetItemReqSODoc(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) GetItemReqSO(c *fiber.Ctx) error {
	idParam := c.Params("sales_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"SalesId": idNum,
	}

	data, status, err := h.Service.GetItemReqSO(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) CreateItemRequest(c *fiber.Ctx) error {
	var body inventory_models.ItemRequestBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateItemRequest(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) UpdateItemRequest(c *fiber.Ctx) error {
	var body inventory_models.ItemRequestBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateItemRequest(&body, nil, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemRequestHandler) DeleteItemRequest(c *fiber.Ctx) error {
	var body inventory_models.ItemRequestBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteItemRequest(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
