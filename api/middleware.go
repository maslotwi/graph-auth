package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// RequireSession validates the Bearer token from the Authorization header.
// Checks Redis first, then falls back to Neo4j and repopulates the cache on success.
// On success it sets "sessionToken" and "email" in request locals for handlers to use.
func RequireSession(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "missing_session"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	email, ok := resolveSessionEmail(token)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session_invalid"})
	}

	c.Locals("sessionToken", token)
	c.Locals("email", email)
	return c.Next()
}
