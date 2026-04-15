package purchasing_handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/purchasing_services"
	"github.com/pierceperado/smpc/utils"
)

func GetPurchaseOrder(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPurchaseOrder(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreatePurchaseOrder(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	fmt.Println("CREATING PO")
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := purchasing_services.CreatePurchaseOrder(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdatePurchaseOrder(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := purchasing_services.UpdatePurchaseOrder(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
