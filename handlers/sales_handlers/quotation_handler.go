package sales_handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/sales_services"
	"github.com/pierceperado/smpc/utils"
)

func GetSalesQuotations(c *fiber.Ctx) error {
	data, status, err := sales_services.GetSalesQuotations(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
func GetLatestQuotations(c *fiber.Ctx) error {
	data, status, err := sales_services.GetLatestQuotations(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
func GetBpiCustomers(c *fiber.Ctx) error {
	data, status, err := sales_services.GetBpiCustomers(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
func GetBpis(c *fiber.Ctx) error {
	data, status, err := sales_services.GetBpis(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
func GetItems(c *fiber.Ctx) error {
	data, status, err := sales_services.GetItems(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetBpi(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := sales_services.GetBpi(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetItem(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		fmt.Println(err)
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := sales_services.GetItem(idNum)
	if err != nil {
		fmt.Println(err)
		return utils.RespondError(c, status, err.Error())
	}

	fmt.Println(data)
	return utils.RespondSuccess(c, data)
}

func GetSalesQuotation(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := sales_services.GetSalesQuotation(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateSalesQuotation(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := sales_services.CreateSalesQuotation(c, tx)
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

// RequestQuotationForEngr - §3.2/§6.3's REQUEST FOR ENGR. action (Phase 4 item 4.1).
func RequestQuotationForEngr(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid quotation id")
	}

	var body struct {
		EngrId uint `json:"engr_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.RequestQuotationForEngr(tx, uint(idNum), body.EngrId, at)
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

func UpdateQuotation(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateFinalizeQuote(c, tx, nil)

	fmt.Print(data)

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
