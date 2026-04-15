package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

// TWO HANDLERS FOR GETTING Book
func GetBooks(c *fiber.Ctx) error {
	data, status, err := setup_services.GetBooks(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// func GetBook(c *fiber.Ctx) error {

// 	idParam := c.Params("id")
// 	idNum, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
// 	}

// 	data, status, err := setup_services.GetBook(idNum)
// 	if err != nil {
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	return utils.RespondSuccess(c, data)
// }

///////////////////////////////

// HANDLER FOR CREATING Book
func CreateBook(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := setup_services.CreateBook(c, tx)
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

// HANDLER FOR UPDATING Book
func UpdateBook(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.UpdateBook(c, tx, nil)
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

// HANDLER DELETE Book
func DeleteBook(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.DeleteBook(c, tx, nil)
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
