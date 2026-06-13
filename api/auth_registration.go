package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RegisterAuthRoutes registers email magic-link registration and verification endpoints.
func RegisterAuthRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")
	auth.Post("/register", HandleRegister)
	auth.Post("/verify", HandleVerify)
}

// HandleRegister sends a magic-link email to the provided address.
//
// @Summary             Register / Request Magic Link
// @Description         Accepts an email address and sends a one-time login link valid for 15 minutes.
// @Tags                Auth
// @Accept              json
// @Produce             json
// @Param               body body object true "JSON body with email field"
// @Success             200 {object} map[string]string "Confirmation that the email was sent"
// @Failure             400 {object} map[string]string "Missing or invalid email"
// @Failure             500 {object} map[string]string "Failed to store token or send email"
// @Router              /api/auth/register [post]
func HandleRegister(c fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.Bind().Body(&body); err != nil || !isValidEmail(body.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_email"})
	}

	token := uuid.NewString()
	if err := storeInRedis("register:"+token, body.Email, 900); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token_store_failed"})
	}

	if err := SendMagicLinkEmail(body.Email, token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "email_send_failed"})
	}

	return c.JSON(fiber.Map{"message": "Check your email for a verification link."})
}

// HandleVerify exchanges a magic-link token for a session token.
//
// @Summary             Verify Magic Link Token
// @Description         Consumes a one-time token from the magic link and returns a session token.
// @Tags                Auth
// @Accept              json
// @Produce             json
// @Param               body body object true "JSON body with token field"
// @Success             200 {object} map[string]interface{} "Session token, email, and root setup flag"
// @Failure             400 {object} map[string]string "Missing token"
// @Failure             401 {object} map[string]string "Token expired or invalid"
// @Failure             500 {object} map[string]string "Failed to create session"
// @Router              /api/auth/verify [post]
func HandleVerify(c fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	if err := c.Bind().Body(&body); err != nil || body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token_required"})
	}

	email, err := getFromRedis("register:" + body.Token)
	if err != nil || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token_expired_or_invalid"})
	}
	_ = deleteFromRedis("register:" + body.Token)

	sessionToken := uuid.NewString()
	if err := storeInRedis("session:"+sessionToken, email, 86400); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session_store_failed"})
	}

	// requiresRootSetup: stub true until Neo4j check is wired up
	return c.JSON(fiber.Map{
		"sessionToken":      sessionToken,
		"email":             email,
		"requiresRootSetup": true,
	})
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
