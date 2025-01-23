package setup_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

func GetPaymentTerms(c *fiber.Ctx) error {
	data, err := setup_services.GetPaymentTerms()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})

}

func GetPaymentTerm(c *fiber.Ctx) error {
	idParam := c.Params("id")

	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err,
		})
	}

	data, status, err := setup_services.GetBrand(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func CreatePaymentTerms(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()                   // Start transactions
	err := setup_services.CreatePaymentTerms(c, tx) // Call functions that will execute the creation of records

	// if error occurs  transaction rollback
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}
	// if error in commit occurs rollback the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"message": "Failed to commit transactions",
		}))

	}

	//return success after commit already done
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}

func UpdatePaymentTerms(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	err := setup_services.UpdatePaymentTerms(c, tx)

	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to updating payment terms",
		})
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"message": "Failed to commit transactions",
		}))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}
func DeletePaymentTerms(c *fiber.Ctx) error {

	tx := initializers.DB.Begin()
	err := setup_services.DeletePaymentTerms(c, tx)

	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed deleting payment terms",
		})
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to commit transaction",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}
