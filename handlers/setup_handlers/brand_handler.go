package setup_handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/setup_services"
)

func GetBrands(c *fiber.Ctx) error {
	data, err := setup_services.GetBrands()

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

func CreateBrand(c *fiber.Ctx) error {
	var body models.Brand
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Cannot bind request",
		})
	}

	tx := initializers.DB.Begin()
	err := setup_services.CreateBrand(tx, &body)

	fmt.Println("Brand At:", body)

	if err != nil {
		fmt.Println("Error:", err)
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed creating brand",
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
		// "data":    data,
	})
}

func UpdateBrand(c *fiber.Ctx) error {

	// var body models.Brand
	// if err := c.BodyParser(&body); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Cannot bind request",
	// 	})
	// }

	// data, err := item_brand_services.Create(body)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
	})
}

func DeleteBrand(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
	})
}
