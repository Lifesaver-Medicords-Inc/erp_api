package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

// TWO HANDLERS FOR GETTING ChartOfAccountS
func GetChartOfAccounts(c *fiber.Ctx) error {

	data, status, err := setup_services.GetChartOfAccounts(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// func GetChartOfAccount(c *fiber.Ctx) error {

// 	idParam := c.Params("id")
// 	idNum, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
// 	}

// 	data, status, err := setup_services.GetChartOfAccount(idNum)
// 	if err != nil {
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }

///////////////////////////////

// HANDLER FOR CREATING ChartOfAccount
// func CreateChartOfAccount(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}
// 	data, status, err := setup_services.CreateChartOfAccount(c, tx)
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

// HANDLER FOR UPDATING ChartOfAccount

// func UpdateChartOfAccount(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := setup_services.UpdateChartOfAccount(c, tx, nil)
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

// HANDLER DELETE ChartOfAccount

// func DeleteChartOfAccount(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := setup_services.DeleteChartOfAccount(c, tx, nil)
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
