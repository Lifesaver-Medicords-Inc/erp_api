package middlewares

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// Login attempt limits. Nothing limited login before, so a password could be guessed
// as fast as the API would answer. Only failed attempts count (SkipSuccessfulRequests),
// so an ordinary sign-in never uses any of the allowance.
//
//   - Per client address and employee id: stops guessing one account's password.
//   - Per client address alone, set higher: stops one source working through many ids.
//
// Both key on the real client address (ClientIP). Keyed on the connection address,
// every visitor through the Cloudflare tunnel would share one allowance, and a single
// guesser could lock the whole company out of login.
//
// The general limiter left commented out in main.go stays off for that same reason,
// and because a page such as Item Entry legitimately fires a dozen requests at once.
var (
	LoginAttemptsPerAccount = limiter.New(limiter.Config{
		Max:                    10,
		Expiration:             5 * time.Minute,
		SkipSuccessfulRequests: true,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "login:" + ClientIP(c) + ":" + loginEmployeeId(c)
		},
		LimitReached: tooManyLoginAttempts,
	})

	LoginAttemptsPerClient = limiter.New(limiter.Config{
		Max:                    30,
		Expiration:             5 * time.Minute,
		SkipSuccessfulRequests: true,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "login:" + ClientIP(c)
		},
		LimitReached: tooManyLoginAttempts,
	})
)

func loginEmployeeId(c *fiber.Ctx) string {
	var body struct {
		EmployeeId string `json:"employee_id"`
	}
	_ = json.Unmarshal(c.Body(), &body)
	return strings.ToLower(strings.TrimSpace(body.EmployeeId))
}

func tooManyLoginAttempts(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"success": false,
		"message": "Too many failed login attempts. Try again in a few minutes.",
	})
}
