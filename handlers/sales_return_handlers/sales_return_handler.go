package sales_return_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/sales_return_services"
	"github.com/pierceperado/smpc/utils"
)

type SalesReturnHandler struct {
	Service *sales_return_services.SalesReturnService
}

func NewSalesReturnHandler(service *sales_return_services.SalesReturnService) *SalesReturnHandler {
	return &SalesReturnHandler{Service: service}
}

func (h *SalesReturnHandler) GetSalesReturn(c *fiber.Ctx) error {
	data, status, err := h.Service.GetSalesReturn(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesReturnHandler) GetSalesReturnById(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid id is required")
	}

	data, status, err := h.Service.GetSalesReturn(map[string]interface{}{"ID": idNum})
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (h *SalesReturnHandler) CreateSalesReturn(c *fiber.Ctx) error {
	var body models.SalesReturnBody

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, code, err := h.Service.CreateSalesReturn(&body, at)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// ApproveSalesReturn signs off on a pending Sales Return. Only a user whose
// Position has been granted the SALES_RETURN_APPROVAL access code (see
// sales_return_services.SalesReturnApprovalAccessCode) can do this - checked
// server-side, not trusted to the client hiding the button.
func (h *SalesReturnHandler) ApproveSalesReturn(c *fiber.Ctx) error {
	salesReturnId, err := strconv.Atoi(c.Params("id"))
	if err != nil || salesReturnId <= 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "a valid sales return id is required")
	}

	status, err := h.Service.ApproveSalesReturn(uint(salesReturnId), actingUserId(c))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}

// actingUserId pulls the numeric user id off the same "at" audit context
// every other write endpoint relies on (see utils/at_util.go and the
// identical helper in item_stock_handlers) - there's no separate
// authentication/session concept in this API beyond that.
func actingUserId(c *fiber.Ctx) uint {
	at, ok := c.Locals("at").(models.At)
	if !ok {
		return 0
	}

	id, err := strconv.Atoi(at.AtUserId)
	if err != nil || id < 0 {
		return 0
	}

	return uint(id)
}
