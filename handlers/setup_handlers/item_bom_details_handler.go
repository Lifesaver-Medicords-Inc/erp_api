package setup_handlers

// import (
// 	"strconv"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/pierceperado/smpc/initializers"
// 	"github.com/pierceperado/smpc/services/setup_services"
// 	"github.com/pierceperado/smpc/utils"
// )

// func GetSetupItemBomDetails(c *fiber.Ctx) error {
// 	data, status, err := setup_services.GetSetupItemBomDetails(nil)
// 	if err != nil {
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }

// func GetSetupItemBomDetail(c *fiber.Ctx) error {
// 	idParam := c.Params("id")
// 	idNum, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
// 	}

// 	conditions := map[string]interface{}{
// 		"id": idNum,
// 	}

// 	data, status, err := setup_services.GetSetupItemBomDetails(conditions)
// 	if err != nil {
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }

// func CreateSetupItemBomDetail(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := setup_services.CreateSetupItemBomDetail(c, tx)
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

// func UpdateSetupItemBomDetail(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := setup_services.UpdateSetupItemBomDetail(c, tx, nil)
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

// func DeleteSetupItemBomDetail(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := setup_services.DeleteSetupItemBomDetail(c, tx, nil)
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
