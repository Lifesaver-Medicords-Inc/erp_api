package setup_handlers

// import (
// 	"strconv"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/pierceperado/smpc/initializers"
// 	"github.com/pierceperado/smpc/services/setup_services"
// 	"github.com/pierceperado/smpc/utils"
// )

// func GetWarehouseAddresses(c *fiber.Ctx) error {
// 	data, status, err := setup_services.GetAddresses(nil)
// 	if err != nil {
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }

// func GetWarehouseAddress(c *fiber.Ctx) error {
// 	idParam := c.Params("id")
// 	idNum, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
// 	}

// 	data, status, err := setup_services.GetWarehouseAddress(idNum)
// 	if err != nil {
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }

// func CreateWarehouseAddress(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := setup_services.CreateWarehouseAddress(c, tx)
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
