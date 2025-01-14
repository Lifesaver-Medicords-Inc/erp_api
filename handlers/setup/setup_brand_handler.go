package setup_brand_handler

import (
	"github.com/gofiber/fiber/v2"
	item_brand_services "github.com/pierceperado/smpc/services/setup"
)

func Get(c *fiber.Ctx) error {

	data, err := item_brand_services.Get()

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})

}

func Create(c *fiber.Ctx) error {

	// var body models.Brand
	// if err := c.BodyParser(&body); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Cannot bind request",
	// 	})
	// }

	// tx := initializers.DB.Begin()
	// var data, err = item_brand_services.Create(tx, body)

	// if err != nil {

	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Failed creating brand",
	// 	})
	// }

	// if err := tx.Commit().Error; err != nil {
	// 	tx.Rollback()
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Failed to commit transaction",
	// 	})
	// }

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		// "data":    data,
	})

}

func Update(c *fiber.Ctx) error {

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

func Delete(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
	})
}
