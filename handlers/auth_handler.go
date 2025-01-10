package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/utils"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var body struct {
		EmployeeId string
		Password   string
		At         models.At
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Cannot bind request",
		})
	}

	var hash []byte
	var err error

	if hash, err = bcrypt.GenerateFromPassword([]byte(body.Password), 10); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal server error",
		})
	}

	user := models.User{EmployeeId: body.EmployeeId, Password: string(hash)}

	if err := initializers.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed creating user",
		})
	}

	userat := models.UserAt{User: user, At: utils.GetAtData(c, body.At)}

	if err := initializers.DB.Create(&userat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed creating userat",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
	})
}
