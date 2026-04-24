package adminhandlers

import (
	"github.com/gofiber/fiber/v2"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type AdminHandler struct {
	Service *adminservices.RedisService
}

func NewAdminHandler(service *adminservices.RedisService) *AdminHandler {
	return &AdminHandler{Service: service}
}

func (h *AdminHandler) ClearAllCache(c *fiber.Ctx) error {
	status, err := h.Service.ClearAllCache(c)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, nil)
}
