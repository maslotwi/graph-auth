package api

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/db"
)

// HandleUserInfo returns OIDC userinfo claims for a valid access token.
//
// @Summary             OIDC UserInfo
// @Description         Returns flat OIDC userinfo claims for a valid RS256 access token. Claims are filtered by the scopes granted to the token.
// @Tags                OAuth2 SSO
// @Produce             json
// @Param               Authorization header string true "Bearer <access_token>"
// @Success             200 {object} UserInfoResponse "OIDC userinfo claims"
// @Failure             401 {object} ErrorResponse "Missing, invalid, or revoked access token"
// @Router              /api/oauth/userinfo [get]
func HandleUserInfo(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_token"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := parseAccessJWT(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_token"})
	}

	email := claims.Subject
	if email == "" {
		email = claims.Email
	}

	profile, err := db.GetRootProfileByEmail(context.Background(), email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_token"})
	}

	response := UserInfoResponse{
		Sub: profile.Email,
	}

	if hasScope(claims.Scopes, ScopeEmail) {
		response.Email = profile.Email
		response.EmailVerified = true
	}

	if hasScope(claims.Scopes, ScopeProfile) {
		response.Name = profile.DisplayName
		response.Picture = profile.Picture
		response.PreferredUsername = preferredUsernameFromEmail(profile.Email)
	}

	return c.JSON(response)
}

func preferredUsernameFromEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	return email[:at]
}
