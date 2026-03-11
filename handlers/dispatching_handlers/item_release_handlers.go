package dispatching_handlers

import (
	"fmt"
	"strconv"

	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/utils"
)

type ItemReleaseHandler struct {
	Service *dispatching_services.ItemReleaseService
}

func NewItemReleaseHandler(service *dispatching_services.ItemReleaseService) *ItemReleaseHandler {
	return &ItemReleaseHandler{Service: service}
}

// GET /item-releases
func (h *ItemReleaseHandler) GetItemReleasesHandler(c *fiber.Ctx) error {
	id := c.Query("id")
	status := c.Query("status")
	salesOrderId := c.Query("salesOrderId")

	conditions := make(map[string]interface{})
	if idNum, _ := strconv.Atoi(id); idNum != 0 {
		conditions["id"] = idNum
	}
	if status != "" {
		conditions["status"] = status
	}
	if soID, _ := strconv.Atoi(salesOrderId); soID != 0 {
		conditions["sales_order_id"] = soID
	}

	data, code, err := h.Service.GetItemReleasesService(conditions)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GET /item-releases/:id
func (h *ItemReleaseHandler) GetItemReleaseHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	conditions := map[string]interface{}{"id": idNum}

	release, code, err := h.Service.GetItemReleaseService(conditions)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, release)
}

// POST /item-releases
func (h *ItemReleaseHandler) CreateItemReleaseHandler(c *fiber.Ctx) error {
	var body models.ItemRelease
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateItemReleaseService(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// PUT /item-releases/:id
func (h *ItemReleaseHandler) UpdateItemReleaseHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	var body models.ItemRelease
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	fmt.Println("Release Body: ", body)

	conditions := map[string]interface{}{
		"id": idNum,
	}
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateItemReleaseService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	fmt.Println("returned data:", data)

	return utils.RespondSuccess(c, data)
}

// DELETE /item-releases/:id
func (h *ItemReleaseHandler) DeleteItemReleaseHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteItemReleaseService(conditions, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemReleaseHandler) GetSalesOrderItemReleaseDetailsHandler(c *fiber.Ctx) error {

	data, code, err := h.Service.GetSalesOrderDetails(nil)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *ItemReleaseHandler) GetItemStockAndLocationsHandler(c *fiber.Ctx) error {
	itemIdParam := c.Params("itemId")
	itemIdNum, err := strconv.ParseUint(itemIdParam, 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Item ID")
	}

	data, code, err := h.Service.GetItemStockAndLocation(uint(itemIdNum))
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
