package purchase_return_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/purchase_return_services"
	"github.com/pierceperado/smpc/utils"
)

type PurchaseReturnHandler struct {
	Service *purchase_return_services.PurchaseReturnService
}

func NewPurchaseReturnHandler(service *purchase_return_services.PurchaseReturnService) *PurchaseReturnHandler {
	return &PurchaseReturnHandler{Service: service}
}

func (h *PurchaseReturnHandler) GetPurchaseReturn(c *fiber.Ctx) error {
	data, status, err := h.Service.GetPurchaseReturn(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PurchaseReturnHandler) GetPurchaseReturnById(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid id is required")
	}

	data, status, err := h.Service.GetPurchaseReturn(map[string]interface{}{"ID": idNum})
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *PurchaseReturnHandler) CreatePurchaseReturn(c *fiber.Ctx) error {
	var body models.PurchaseReturnBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreatePurchaseReturn(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
