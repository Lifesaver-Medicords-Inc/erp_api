package sales_quotation_handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
)

func Get(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
	})
}

func Create(c *fiber.Ctx) error { 
	
	var body struct {
		FirstName  string    `json:"first_name"`
		LastName   string    `json:"last_name"`
		Department string    `json:"department"`
		Position   string    `json:"position"`
		At         models.At `json:"at"`
	}

	if err := c.BodyParser(&body); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Cannot bind request",
		})
	} 

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": body,
	})

}

func Update(c *fiber.Ctx) error {
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
