package setup_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

type FixedAssetHandler struct {
	Service *setup_services.FixedAssetService
}

func NewFixedAssetHandler(service *setup_services.FixedAssetService) *FixedAssetHandler {
	return &FixedAssetHandler{Service: service}
}

func (h *FixedAssetHandler) GetFixedAssets(c *fiber.Ctx) error {
	search := c.Query("search")

	var id int
	if idParam := c.Query("id"); idParam != "" {
		var err error
		id, err = strconv.Atoi(idParam)
		if err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "invalid id")
		}
	}

	data, status, pagination, err := h.Service.GetFixedAssets(nil, search, id)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
}

func (h *FixedAssetHandler) GetFixedAsset(c *fiber.Ctx) error {
	idNum, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := h.Service.GetFixedAsset(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *FixedAssetHandler) CreateFixedAsset(c *fiber.Ctx) error {
	var body accounting_models.FixedAsset
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateFixedAsset(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *FixedAssetHandler) UpdateFixedAsset(c *fiber.Ctx) error {
	var body accounting_models.FixedAsset
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateFixedAsset(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *FixedAssetHandler) DeleteFixedAsset(c *fiber.Ctx) error {
	var body accounting_models.FixedAsset
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteFixedAsset(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
