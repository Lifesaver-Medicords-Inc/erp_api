package public_handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/middlewares"
	"github.com/pierceperado/smpc/services/public_services"
	"github.com/pierceperado/smpc/utils"
)

func CreateAccount(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := public_services.CreateAccount(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func LoginAccount(c *fiber.Ctx) error {
	data, status, err := public_services.LoginAccount(c)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func LogoutAccount(c *fiber.Ctx) error {
	// Ends the session on the server. Logout used to clear only a cookie the desktop
	// apps never send back, so the token itself stayed valid for the rest of its 24
	// hours. A missing or invalid token has nothing to revoke, and logout still
	// succeeds.
	token := strings.TrimSpace(c.Get("Authorization"))
	if token == "" {
		token = c.Cookies("Authorization")
	}
	if token != "" {
		if err := middlewares.RevokeToken(token); err != nil {
			log.Println("logout: token not revoked:", err)
		}
	}

	public_services.LogoutAccount(c)

	return utils.RespondSuccess(c, nil)
}
