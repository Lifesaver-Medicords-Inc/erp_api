package setup_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
)

func GetUnitMeasurements(c *fiber.Ctx) error {
	data, err := setup_services.GetUnitMeasurements()

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

func CreateUnitMeasurement(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()                      // Start transactions
	err := setup_services.CreateUnitMeasurement(c, tx) // Call functions that will execute the creation of records

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

func UpdateUnitMeasurement(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	err := setup_services.UpdateUnitMeasurment(c, tx)

	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to updating brand",
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
func DeleteUnitMeasurement(c *fiber.Ctx) error {

	tx := initializers.DB.Begin()
	err := setup_services.DeleteUnitMeasurment(c, tx)

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
