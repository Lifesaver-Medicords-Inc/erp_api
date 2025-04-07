package purchasing_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/purchasing_services"
	"github.com/pierceperado/smpc/utils"
)

func GetPurchasingList(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPurchasingList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func GetPurchasingListSupplier(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPurchasingListSupplier(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}