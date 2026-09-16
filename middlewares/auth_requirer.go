package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/utils"
)

// RequireAuth used to print the token, its claims and the user on every request. That
// put a live, reusable session token on the console for every call any app made -
// anyone who could see the window, or a copy of its output, could act as that user
// until the token expired. The prints are gone; the checks are unchanged apart from
// where the token may come from (tokenFromRequest) and the logout check.
func RequireAuth(c *fiber.Ctx) error {
	tokenString := tokenFromRequest(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Missing authentication token",
		})
	}

	token, err := parseToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid authentication token",
		})
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "Expired authentication token",
				})
			}
		}

		// Logged out before it expired (RevokeToken).
		if isTokenRevoked(tokenString) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Revoked authentication token",
			})
		}

		userID, ok := claims["sub"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID",
			})
		}

		if userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Missing user ID",
			})
		}

		var userObj models.User
		if count := initializers.DB.First(&userObj, "id = ?", userID).RowsAffected; count == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}

		atMap, ok := claims["at"].(map[string]interface{})
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid at data",
			})
		}

		var atObj models.At
		if err := utils.MapToStruct(atMap, &atObj); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid at data",
			})
		}

		c.Locals("user", userObj)
		c.Locals("at", atObj)

		return c.Next()
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success": false,
		"message": "Invalid authentication token",
	})
}
