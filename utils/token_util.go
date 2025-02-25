package utils

import (
	"errors"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pierceperado/smpc/models"
)

func CreateAuthToken(c *fiber.Ctx, at models.At, uid uint) error {
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"at":  GetAtData(c, at),
	})

	secretKey := os.Getenv("SECRET_KEY")
	tokenString, err := tokenClaims.SignedString([]byte(secretKey))
	if err != nil {
		return errors.New("failed creating token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: fiber.CookieSameSiteLaxMode,
		HTTPOnly: true,
		Secure:   true,
	})

	return nil
}
