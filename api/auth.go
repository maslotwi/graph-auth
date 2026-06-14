package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/db"
)

const (
	authCodeTTLSeconds    = 60
	accessTokenTTL        = 15 * time.Minute
	accessTokenTTLSeconds = 900
)

var internalOAuthScopes = map[string]struct{}{
	ScopeFertile: {},
	ScopeClients: {},
}

// RegisterOAuthRoutes registers the central SSO engine routes.
func RegisterOAuthRoutes(app *fiber.App) {
	oauth := app.Group("/api/oauth")

	oauth.Post("/authorize", RequireSession, HandleAuthorize)
	oauth.Post("/token", HandleToken)
}

// HandleAuthorize issues an OAuth2 authorization code for an authenticated session.
//
// @Summary             Authorize SSO Request
// @Description         Creates a short-lived authorization code for the authenticated user and returns the client redirect URL. Requires a valid session with the clients scope excluded from granted OAuth scopes.
// @Tags                OAuth2 SSO
// @Accept              json
// @Produce             json
// @Param               Authorization header string true "Bearer <token> where token is a Session.token"
// @Param               body body AuthorizeRequest true "OAuth2 client and redirect parameters"
// @Success             200 {object} AuthorizeResponse "Returns the callback URL containing the auth code"
// @Failure             400 {object} ErrorResponse "Invalid request or unknown client"
// @Failure             401 {object} ErrorResponse "Unauthorized due to missing or invalid session"
// @Failure             403 {object} ErrorResponse "Requested scope exceeds session scopes"
// @Failure             500 {object} ErrorResponse "Failed to store authorization code"
// @Router              /api/oauth/authorize [post]
func HandleAuthorize(c fiber.Ctx) error {
	sessionToken := c.Locals("sessionToken").(string)

	var body AuthorizeRequest
	if err := c.Bind().Body(&body); err != nil || body.ClientID == "" || body.RedirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	if _, err := db.GetClientByID(context.Background(), body.ClientID); err != nil {
		if errors.Is(err, db.ErrClientNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "unauthorized_client"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_lookup_failed"})
	}

	sessionScopes, active, err := db.ActiveSessionScopes(context.Background(), sessionToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_lookup_failed"})
	}
	if !active {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session_invalid"})
	}

	grantedScopes := filterOAuthScopes(sessionScopes)
	if body.Scope != "" {
		requested := normalizeScopes(strings.Fields(body.Scope))
		if !isSubsetScopes(requested, grantedScopes) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "invalid_scope"})
		}
		grantedScopes = requested
	}

	authCode := "code_" + generateSecureUUID()
	payload, err := json.Marshal(authCodePayload{
		Email:    c.Locals("email").(string),
		ClientID: body.ClientID,
		Scopes:   grantedScopes,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "serialization_failure"})
	}

	if err := storeInRedis("code:"+authCode, string(payload), authCodeTTLSeconds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "code_store_failed"})
	}

	redirectTo := fmt.Sprintf("%s?code=%s&state=%s", body.RedirectURI, authCode, body.State)
	return c.JSON(AuthorizeResponse{
		Status:     "success",
		RedirectTo: redirectTo,
	})
}

// HandleToken exchanges an authorization code for a signed JWT access token.
//
// @Summary             Exchange Auth Code for JWT
// @Description         Consumes a short-lived authorization code and issues a signed JWT access token with a Redis-backed jti nonce for revocation.
// @Tags                OAuth2 SSO
// @Accept              json
// @Produce             json
// @Param               body body TokenExchangeRequest true "JSON body containing grant_type, code, client_id, and client_secret"
// @Success             200 {object} TokenExchangeResponse "Returns the standard OAuth2 JWT access token payload"
// @Failure             400 {object} ErrorResponse "Invalid request format or unsupported grant type"
// @Failure             401 {object} ErrorResponse "Invalid client credentials or authorization code"
// @Failure             500 {object} ErrorResponse "Failed to mint access token"
// @Router              /api/oauth/token [post]
func HandleToken(c fiber.Ctx) error {
	var body TokenExchangeRequest
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	if body.GrantType != "" && body.GrantType != "authorization_code" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "unsupported_grant_type"})
	}
	if body.Code == "" || body.ClientID == "" || body.ClientSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	valid, err := db.VerifyClientCredentials(context.Background(), body.ClientID, body.ClientSecret)
	if err != nil {
		if errors.Is(err, db.ErrClientNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_client"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_lookup_failed"})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_client"})
	}

	payloadRaw, err := consumeFromRedis("code:" + body.Code)
	if err != nil || payloadRaw == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_grant"})
	}

	var payload authCodePayload
	if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_grant"})
	}
	if payload.ClientID != body.ClientID {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_grant"})
	}

	accessToken, err := mintAccessJWT(payload.Email, body.ClientID, payload.Scopes, accessTokenTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "token_mint_failed"})
	}

	return c.JSON(TokenExchangeResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   accessTokenTTLSeconds,
		Scopes:      payload.Scopes,
	})
}

func filterOAuthScopes(scopes []string) []string {
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if _, internal := internalOAuthScopes[scope]; internal {
			continue
		}
		out = append(out, scope)
	}
	return out
}
