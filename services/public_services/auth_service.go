package public_services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
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

	employeeId := utils.GenerateEmployeeId(body.Department, body.Position, user.ID)
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
	fmt.Println("LOGIN ACCOUNT")
	var body models.UserAt
	if err := c.BodyParser(&body); err != nil {
		return user, fiber.StatusBadRequest, errors.New("cannot bind request")
	}
	fmt.Println("LOGIN ACCOUNT33333")

	conditions := map[string]interface{}{
		"employee_id": body.EmployeeId,
	}

	if err := services.DbGet(&user, conditions); err != nil {
		return user, fiber.StatusUnauthorized, errors.New("invalid user credential")
	}

	fmt.Println("LOGIN ACCOUNT4444", user.ID)

	if err := utils.CompareUserPassword(user.Password, body.Password); err != nil {
		return user, fiber.StatusUnauthorized, errors.New("invalid user credential")
	}
	fmt.Println("LOGIN ACCOUNT555")

	body.At.AtUserId = strconv.Itoa(int(user.ID))
	if err := utils.CreateAuthToken(c, body.At, user.ID); err != nil {
		return user, fiber.StatusUnauthorized, err
	}
	fmt.Println("LOGIN ACCOUNT6666")

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
