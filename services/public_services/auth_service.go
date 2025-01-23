package public_services

import (
	"errors"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateAccount(c *fiber.Ctx, tx *gorm.DB) (models.User, int, error) {
	var user models.User

	var body models.UserAt
	if err := c.BodyParser(&body); err != nil {
		return user, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	user = models.User{UserContent: models.UserContent{FirstName: body.FirstName, LastName: body.LastName, Department: body.Department, Position: body.Position}}
	if err := services.DbInsert(tx, &user); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed creating user")
	}

	employeeId := utils.GenerateEmployeeId(body.Department, body.Position, user.ID)
	hash, err := bcrypt.GenerateFromPassword([]byte(employeeId), 10)
	if err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed generating password")
	}

	user.EmployeeId = employeeId
	user.Password = string(hash)
	if err := services.DbUpdate(tx, &user, nil); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed udpating user")
	}

	userat := models.UserAt{RefId: user.ID, EmployeeId: employeeId, UserContent: models.UserContent{FirstName: body.FirstName, LastName: body.LastName, Department: body.Department, Position: body.Position, Password: user.Password}, At: utils.GetAtData(c, body.At)}
	if err := services.DbInsert(tx, &userat); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed creating user")
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

	if err := services.DbGet(&user, conditions); err != nil {
		return user, fiber.StatusUnauthorized, errors.New("invalid user credential")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return user, fiber.StatusUnauthorized, errors.New("invalid user credential")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"at":  utils.GetAtData(c, body.At),
	})

	secretKey := os.Getenv("SECRET_KEY")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed creating token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: fiber.CookieSameSiteLaxMode,
		HTTPOnly: true,
		Secure:   true,
	})

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
