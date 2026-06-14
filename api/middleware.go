package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// RequireSession validates the Bearer token from the Authorization header against Redis.
// On success it sets "sessionToken" and "email" in request locals for handlers to use.
func RequireSession(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "missing_session"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	email, err := getFromRedis("session:" + token)
	if err != nil || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session_invalid"})
	}

	c.Locals("sessionToken", token)
	c.Locals("email", email)
	return c.Next()
}
