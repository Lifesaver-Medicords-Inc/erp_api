package public_services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

func CreateAccount(c *fiber.Ctx, tx *gorm.DB) (models.User, int, error) {
	var user models.User

	var body models.UserAt

	if err := c.BodyParser(&body); err != nil {
		return user, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	user = models.User{UserContent: body.UserContent}
	if err := services.DbInsert(tx, &user); err != nil {
		fmt.Println("Error:", err)
		return user, fiber.StatusInternalServerError, errors.New("failed creating user")
	}

	employeeId := utils.GenerateEmployeeId(body.Department, user.Position.Name, user.ID)
	user.EmployeeId = employeeId

	password, err := utils.GenerateUserPassword(employeeId)
	if err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed generating password")
	}
	user.Password = password

	if err := services.DbUpdate(tx, &user, nil); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed udpating user")
	}

	body.RefId = user.ID
	body.EmployeeId = employeeId
	body.Password = password
	body.At = utils.GetAtData(c, body.At)

	if err := services.DbInsert(tx, &body); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	return user, 0, nil
}

func LoginAccount(c *fiber.Ctx) (models.User, int, error) {
	var user models.User

	var body models.UserAt
	if err := c.BodyParser(&body); err != nil {
		return user, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	conditions := map[string]interface{}{
		"employee_id": body.EmployeeId,
	}

	if err := initializers.DB.
		Preload("Position").
		Where(conditions).
		First(&user).Error; err != nil {
		return user, fiber.StatusUnauthorized, errors.New("Invalid user employee id")
	}

	if err := utils.CompareUserPassword(user.Password, body.Password); err != nil {
		return user, fiber.StatusUnauthorized, errors.New("Invalid user password")
	}

	body.AtUserId = strconv.Itoa(int(user.ID))
	if err := utils.CreateAuthToken(c, body.At, user.ID); err != nil {
		return user, fiber.StatusUnauthorized, err
	}
	fmt.Println("LOGIN:", user.Position.Name)

	return user, 0, nil
}

func LogoutAccount(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		SameSite: fiber.CookieSameSiteLaxMode,
		HTTPOnly: true,
		Secure:   true,
	})
}
