package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pierceperado/smpc/initializers"
)

// ClientIP returns the address a request really came from.
//
// Everything arriving through the Cloudflare tunnel reaches the API from cloudflared
// on this same machine, so the connection address is loopback for all of it and the
// request log could not tell one visitor from another. cloudflared passes the
// visitor's address in Cf-Connecting-Ip. That header is believed only when the
// connection itself is loopback - from anywhere else a client could simply send one -
// which is why this is not Fiber's ProxyHeader setting: that would believe the header
// from everyone, and change c.IP() for every existing caller besides.
func ClientIP(c *fiber.Ctx) string {
	remote := c.Context().RemoteIP()
	if remote.IsLoopback() {
		if forwarded := strings.TrimSpace(c.Get("Cf-Connecting-Ip")); net.ParseIP(forwarded) != nil {
			return forwarded
		}
	}
	return remote.String()
}

// tokenFromRequest reads the session token from the Authorization header.
//
// The ?Authorization= query form is accepted only on a WebSocket upgrade, because the
// red-box sockets in inventory and engineering send it that way. On an ordinary
// request a token in the URL is written into tunnel and proxy logs, so it is no longer
// read there.
func tokenFromRequest(c *fiber.Ctx) string {
	if token := strings.TrimSpace(c.Get("Authorization")); token != "" {
		return token
	}
	if websocket.IsWebSocketUpgrade(c) {
		return strings.TrimSpace(c.Query("Authorization"))
	}
	return ""
}

// parseToken checks the token's signature against SECRET_KEY.
func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", method)
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
}

const revokedTokenKeyPrefix = "auth:revoked:"

// Stored hashed: the key only has to recognise the token, and a readable copy in
// Redis would be as usable as the token itself.
func revokedTokenKey(tokenString string) string {
	sum := sha256.Sum256([]byte(tokenString))
	return revokedTokenKeyPrefix + hex.EncodeToString(sum[:])
}

// RevokeToken ends a session before its 24-hour expiry. Logout used to clear only a
// cookie the desktop apps never send back, so a logged-out token - or one copied off
// a console or a log - kept working until it expired. The entry lives exactly as long
// as the token would have, so nothing accumulates.
func RevokeToken(tokenString string) error {
	token, err := parseToken(tokenString)
	if err != nil || !token.Valid {
		return errors.New("invalid authentication token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid authentication token")
	}

	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return errors.New("authentication token has no expiry")
	}

	ttl := time.Until(exp.Time)
	if ttl <= 0 {
		return nil
	}

	return initializers.RC.Set(context.Background(), revokedTokenKey(tokenString), 1, ttl).Err()
}

// isTokenRevoked fails open: when Redis cannot be reached the request is let through
// and the failure logged. Redis is already required at startup, and failing closed
// would sign every user out of every app for the length of an outage.
func isTokenRevoked(tokenString string) bool {
	n, err := initializers.RC.Exists(context.Background(), revokedTokenKey(tokenString)).Result()
	if err != nil {
		log.Println("auth: could not check token revocation:", err)
		return false
	}
	return n > 0
}
