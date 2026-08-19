package adminhandlers

import (
	"github.com/gofiber/fiber/v2"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type AccessModuleHandler struct {
	AccessModuleService *adminservices.AccessModuleService
}

func NewAccessModuleHandler(service *adminservices.AccessModuleService) *AccessModuleHandler {
	return &AccessModuleHandler{
		AccessModuleService: service,
	}
}

// GetAllAccessModulesHandler serves the full access-code catalog (read-only) - the
// Admin Access Control screen's tree is built entirely from this. Granting/revoking
// still goes through the existing /api/position-access endpoints (tbl_position_access
// itself is unchanged); this only tells the UI what codes exist to check.
func (a *AccessModuleHandler) GetAllAccessModulesHandler(c *fiber.Ctx) error {
	data, status, err := a.AccessModuleService.GetAllAccessModulesService()
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
