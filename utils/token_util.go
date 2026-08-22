package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pierceperado/smpc/models"
)

func CreateAuthToken(c *fiber.Ctx, at models.At, uid uint) error {
	fmt.Println("UID", uid)
	fmt.Println("UIDat", at)

	// GetAtData rebuilds AtUserId from c.Locals("user") - but this is called from
	// LoginAccount, before RequireAuth has ever run on this request, so
	// c.Locals("user") is never set yet and GetAtData silently baked AtUserId="0"
	// into every login token for every user. That "at" claim is copied verbatim
	// into c.Locals("at") on every later request (auth_requirer.go), so
	// actingUserId(c) (used for e.g. reservation approval) always resolved to 0
	// and got rejected by the userId==0 guard - regardless of what was actually
	// granted in tbl_position_access. `uid` is the real, just-authenticated user
	// id passed in by the caller, so use it here instead of trusting Locals.
	atData := GetAtData(c, at)
	atData.AtUserId = strconv.Itoa(int(uid))

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"at":  atData,
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
