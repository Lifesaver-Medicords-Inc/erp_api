package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/utils"
)

// ManageUsersAccessCode is the tbl_position_access code behind the Admin app's Users
// screen. Creating, editing, deleting and re-positioning users, and setting their
// passwords, all require it - the same grant that shows the screen, checked here so a
// hidden button is not the only thing in the way.
const ManageUsersAccessCode = "ADMIN USERS"

// RequireAccessCode lets a request through only when the signed-in user's position holds
// code. The position is read off the user RequireAuth loaded from the database on this
// same request, so a position changed since login applies at once. Register it after
// RequireAuth.
func RequireAccessCode(code string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok || user.ID == 0 {
			return utils.RespondError(c, fiber.StatusUnauthorized, "Missing authentication token")
		}

		var count int64
		if err := initializers.DB.Raw(
			`SELECT COUNT(*) FROM tbl_position_access WHERE position_id = ? AND LTRIM(RTRIM(code)) = ?`,
			user.PositionId, code,
		).Scan(&count).Error; err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "failed checking access")
		}

		if count == 0 {
			return utils.RespondError(c, fiber.StatusForbidden, "Your position does not have "+code+" access")
		}

		return c.Next()
	}
}
