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

	user := models.User{FirstName: body.FirstName, LastName: body.LastName, Department: body.Department, Position: body.Position}

	if err := initializers.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed creating user",
		})
	}

	employeeId := utils.GenerateEmployeeId(user.Department, user.Position, user.ID)

	var hash []byte
	var err error

	if hash, err = bcrypt.GenerateFromPassword([]byte(employeeId), 10); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal server error",
		})
	}

	user.EmployeeId = employeeId
	user.Password = string(hash)

	if err := initializers.DB.Model(&user).Updates(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed updating user",
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
