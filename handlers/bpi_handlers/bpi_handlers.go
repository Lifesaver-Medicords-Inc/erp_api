package bpi_handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"

	"github.com/pierceperado/smpc/services/bpi_services"
	"github.com/pierceperado/smpc/utils"
)

func GetBpis(c *fiber.Ctx) error {
	data, status, err := bpi_services.GetBpis(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func GetBpiUsers(c *fiber.Ctx) error {
	fmt.Println("FMT GET BPI USERS")
	idParam := c.Params("employee_id")

	data, status, err := bpi_services.GetBpiUsers(idParam)
	fmt.Println("DATA USER", data)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetBpiEntityRecords(c *fiber.Ctx) error {
	data, status, err := bpi_services.GetBpiEntityRecords(nil)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetBpiItemList(c *fiber.Ctx) error {
	data, status, err := bpi_services.GetBpiItemList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetBpiItemListPaged serves BPI's Add Item picker - 20 rows a page, searched server-side.
// ?search= and ?page= (1-based); the response's pagination carries page/total_pages.
func GetBpiItemListPaged(c *fiber.Ctx) error {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))

	data, pagination, status, err := bpi_services.GetBpiItemListPaged(search, page)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
}

func CreateBpi(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transactions")
	}

	data, status, err := bpi_services.CreateBpi(c, tx)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transactions")
	}
	return utils.RespondSuccess(c, data)
}

func CreateBpiParentFromBranch(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transactions")
	}

	data, status, err := bpi_services.CreateBpiParentFromBranch(c, tx)

	fmt.Println("DATA HANDLER", data)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transactions")
	}
	return utils.RespondSuccess(c, data)
}

func UpdateBpi(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transactions")
	}
	data, status, err := bpi_services.UpdateBpi(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transactions")
	}

	return utils.RespondSuccess(c, data)
}

func UpdateBpiMainBranch(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed to start transaction")
	}

	// Call service
	data, status, err := bpi_services.UpdateBpiMainBranch(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}
