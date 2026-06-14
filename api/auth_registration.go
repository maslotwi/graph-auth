package api

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/maslotwi/graph-auth/db"
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
// @Param               body body RegisterRequest true "JSON body with email field"
// @Success             200 {object} RegisterResponse "Confirmation that the email was sent"
// @Failure             400 {object} ErrorResponse "Missing or invalid email"
// @Failure             500 {object} ErrorResponse "Failed to store token or send email"
// @Router              /api/auth/register [post]
func HandleRegister(c fiber.Ctx) error {
	var body RegisterRequest
	if err := c.Bind().Body(&body); err != nil || !isValidEmail(body.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_email"})
	}

	token := uuid.NewString()
	if err := storeInRedis("register:"+token, body.Email, 900); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "token_store_failed"})
	}

	if err := SendMagicLinkEmail(body.Email, token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "email_send_failed"})
	}

	return c.JSON(RegisterResponse{Message: "Check your email for a verification link."})
}

// HandleVerify exchanges a magic-link token for a session token.
//
// @Summary             Verify Magic Link Token
// @Description         Consumes a one-time token from the magic link and returns a session token.
// @Tags                Auth
// @Accept              json
// @Produce             json
// @Param               body body VerifyRequest true "JSON body with token, name, and scopes fields"
// @Success             200 {object} VerifyResponse "Session token and email"
// @Failure             400 {object} ErrorResponse "Missing token"
// @Failure             401 {object} ErrorResponse "Token expired or invalid"
// @Failure             500 {object} ErrorResponse "Failed to create session"
// @Router              /api/auth/verify [post]
func HandleVerify(c fiber.Ctx) error {
	var body VerifyRequest
	if err := c.Bind().Body(&body); err != nil || body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "token_required"})
	}

	email, err := getFromRedis("register:" + body.Token)
	if err != nil || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "token_expired_or_invalid"})
	}
	_ = deleteFromRedis("register:" + body.Token)

	deviceName := body.Name
	if deviceName == "" {
		deviceName = "Primary Device"
	}

	sessionToken := uuid.NewString()
	deviceSession := db.Session{
		Token:      sessionToken,
		DeviceName: deviceName,
		Scopes:     withScope(normalizeScopes(body.Scopes), ScopeFertile),
		IsActive:   true,
	}

	if err := db.CreateSessionFromRoot(context.Background(), email, deviceSession); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_create_failed"})
	}

	if err := storeInRedis("session:"+sessionToken, email, sessionCacheTTLSeconds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_store_failed"})
	}

	return c.JSON(VerifyResponse{
		SessionToken: sessionToken,
		Email:        email,
	})
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
