package job_orders_handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/job_order_services"
	"github.com/pierceperado/smpc/utils"
)

func GetJobOrder(c *fiber.Ctx) error {
	fmt.Println("FMT GET JOB ORDER")

	idParam := c.Params("user_id")

	userId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid user_id")
	}

	data, status, err := job_order_services.GetJobOrder(userId)
	fmt.Println("JOB ORDER", data)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetSalesJobOrder(c *fiber.Ctx) error {
	fmt.Println("FMT GET JOB ORDER SALES")

	salesOrder := c.Params("sales_order")

	data, status, err := job_order_services.GetSalesJobOrder(salesOrder)
	fmt.Println("JOB ORDER (SALES)", data)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetSalesDetailsJobOrder(c *fiber.Ctx) error {
	fmt.Println("FMT GET JOB ORDER SALES")

	idParam := c.Params("order_id")

	orderId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid order_id")
	}

	data, status, err := job_order_services.GetSalesDetailsJobOrder(orderId)
	fmt.Println("JOB ORDER (SALES DETAILS)", data)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetComponents(c *fiber.Ctx) error {
	fmt.Println("FMT GET COMPONENTS")

	idParam := c.Params("bom_id")

	bomId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid order_id")
	}

	data, status, err := job_order_services.GetComponents(bomId)
	fmt.Println("ITEM COMPONENTS", data)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func CreateJobOrder(c *fiber.Ctx) error {

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := job_order_services.CreateJobOrder(c, tx)
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

func UpdateJobOrder(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := job_order_services.UpdateJobOrder(c, tx, nil)
	fmt.Println("JOB ORDER DATA: ", data)
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
