package setup_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

type AssetCategoryHandler struct {
	Service *setup_services.AssetCategoryService
}

func NewAssetCategoryHandler(service *setup_services.AssetCategoryService) *AssetCategoryHandler {
	return &AssetCategoryHandler{Service: service}
}

func (h *AssetCategoryHandler) GetAssetCategories(c *fiber.Ctx) error {
	search := c.Query("search")

	var id int
	if idParam := c.Query("id"); idParam != "" {
		var err error
		id, err = strconv.Atoi(idParam)
		if err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "invalid id")
		}
	}

	data, status, pagination, err := h.Service.GetAssetCategories(nil, search, id)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
}

func (h *AssetCategoryHandler) GetAssetCategory(c *fiber.Ctx) error {
	idNum, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := h.Service.GetAssetCategory(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *AssetCategoryHandler) CreateAssetCategory(c *fiber.Ctx) error {
	var body accounting_models.AssetCategory
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateAssetCategory(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *AssetCategoryHandler) UpdateAssetCategory(c *fiber.Ctx) error {
	var body accounting_models.AssetCategory
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.UpdateAssetCategory(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *AssetCategoryHandler) DeleteAssetCategory(c *fiber.Ctx) error {
	var body accounting_models.AssetCategory
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.DeleteAssetCategory(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
