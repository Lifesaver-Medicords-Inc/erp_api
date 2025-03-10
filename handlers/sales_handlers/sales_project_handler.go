package sales_handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/sales_services"
	"github.com/pierceperado/smpc/utils"
)

func GetSalesProject(c *fiber.Ctx) error {
	data, status, err := sales_services.GetSalesProjects(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func CreateSalesProject(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := sales_services.CreateSalesProject(c, tx)

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

func CreateItemSetTab(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := sales_services.CreateNewItems(c, tx)

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

// func UpdateProjectMultiplier(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := sales_services.UpdateSalesProjectMultiplier()
// 	if err != nil {
// 		tx.Rollback()
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		tx.Rollback()
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
// 	}

// 	return utils.RespondSuccess(c, data)
// }

func UpdateProjectCondition(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateProjectAdvancedCondition(c, tx, nil)

	fmt.Println("Update Body:", data)

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

func UpdateProjectContent(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateProjectContents(c, tx, nil)

	fmt.Println("Update Body:", data)

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
