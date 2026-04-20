package bin_location_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/bin_location_services"
	"github.com/pierceperado/smpc/utils"
)

type BinLocationHandler struct {
	Service *bin_location_services.BinLocationService
}

func NewBinLocationHandler(service *bin_location_services.BinLocationService) *BinLocationHandler {
	return &BinLocationHandler{Service: service}
}

func (h *BinLocationHandler) GetBinLocations(c *fiber.Ctx) error {
	idParam := c.Params("item_id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	conditions := map[string]interface{}{
		"ItemId": idNum,
	}

	data, status, err := h.Service.GetBinLocations(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
