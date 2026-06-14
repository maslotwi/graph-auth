package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/db"
	"github.com/maslotwi/graph-auth/helpers/environment"
)

const (
	authCodeTTLSeconds    = 60
	accessTokenTTL        = 15 * time.Minute
	accessTokenTTLSeconds = 900
)

var errUnsupportedContentType = errors.New("unsupported content type")

// RegisterOAuthRoutes registers the central SSO engine routes.
func RegisterOAuthRoutes(app *fiber.App) {
	oauth := app.Group("/api/oauth")

	oauth.Get("/authorize", HandleAuthorizeRedirect)
	oauth.Post("/authorize", RequireSession, HandleAuthorize)
	oauth.Post("/token", HandleToken)
	oauth.Get("/userinfo", HandleUserInfo)
}

// HandleAuthorizeRedirect starts the browser OAuth flow by redirecting to the consent page.
//
// @Summary             Start OAuth Authorization
// @Description         Validates the client and redirect URI, then redirects the browser to the frontend consent page.
// @Tags                OAuth2 SSO
// @Produce             json
// @Param               client_id query string true "OAuth2 Client ID"
// @Param               redirect_uri query string true "Registered OAuth2 Redirect URI"
// @Param               state query string false "OAuth2 State string"
// @Param               scope query string false "Space-delimited OAuth scopes"
// @Param               response_type query string false "OAuth response type, must be code when provided"
// @Success             302 {string} string "Redirect to the frontend consent page"
// @Failure             400 {object} ErrorResponse "Invalid request, unknown client, or unregistered redirect URI"
// @Router              /api/oauth/authorize [get]
func HandleAuthorizeRedirect(c fiber.Ctx) error {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	state := c.Query("state")
	scope := c.Query("scope")
	responseType := c.Query("response_type")

	if clientID == "" || redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}
	if responseType != "" && responseType != "code" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "unsupported_response_type"})
	}

	client, err := db.GetClientByID(context.Background(), clientID)
	if err != nil {
		if errors.Is(err, db.ErrClientNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "unauthorized_client"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_lookup_failed"})
	}
	if !db.ClientHasRedirectURI(client, redirectURI) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_redirect_uri"})
	}

	consentURL := fmt.Sprintf(
		"%s/sso/consent?client_id=%s&redirect_uri=%s&state=%s&scope=%s",
		strings.TrimRight(environment.FrontendURL, "/"),
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
		url.QueryEscape(scope),
	)
	return c.Redirect().Status(fiber.StatusFound).To(consentURL)
}

// HandleAuthorize issues an OAuth2 authorization code for an authenticated session.
//
// @Summary             Authorize SSO Request
// @Description         Creates a short-lived authorization code for the authenticated user and returns the client redirect URL. Requires a registered redirect URI and valid OAuth scopes.
// @Tags                OAuth2 SSO
// @Accept              json
// @Produce             json
// @Param               Authorization header string true "Bearer <token> where token is a Session.token"
// @Param               body body AuthorizeRequest true "OAuth2 client and redirect parameters"
// @Success             200 {object} AuthorizeResponse "Returns the callback URL containing the auth code"
// @Failure             400 {object} ErrorResponse "Invalid request, unknown client, or unregistered redirect URI"
// @Failure             401 {object} ErrorResponse "Unauthorized due to missing or invalid session"
// @Failure             403 {object} ErrorResponse "Invalid OAuth scope"
// @Failure             500 {object} ErrorResponse "Failed to store authorization code"
// @Router              /api/oauth/authorize [post]
func HandleAuthorize(c fiber.Ctx) error {
	var body AuthorizeRequest
	if err := c.Bind().Body(&body); err != nil || body.ClientID == "" || body.RedirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	client, err := db.GetClientByID(context.Background(), body.ClientID)
	if err != nil {
		if errors.Is(err, db.ErrClientNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "unauthorized_client"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_lookup_failed"})
	}
	if !db.ClientHasRedirectURI(client, body.RedirectURI) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_redirect_uri"})
	}

	grantedScopes, ok := resolveOAuthScopes(body.Scope)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "invalid_scope"})
	}

	authCode := "code_" + generateSecureUUID()
	payload, err := json.Marshal(authCodePayload{
		Email:       c.Locals("email").(string),
		ClientID:    body.ClientID,
		RedirectURI: body.RedirectURI,
		Scopes:      grantedScopes,
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
// @Description         Consumes a short-lived authorization code and issues a signed RS256 JWT access token with a Redis-backed jti nonce for revocation. Accepts application/json or application/x-www-form-urlencoded bodies and client credentials via body or HTTP Basic auth.
// @Tags                OAuth2 SSO
// @Accept              json
// @Accept              mpfd
// @Produce             json
// @Param               body body TokenExchangeRequest true "Token exchange parameters"
// @Success             200 {object} TokenExchangeResponse "Returns the standard OAuth2 JWT access token payload"
// @Failure             400 {object} ErrorResponse "Invalid request format or unsupported grant type"
// @Failure             401 {object} ErrorResponse "Invalid client credentials or authorization code"
// @Failure             500 {object} ErrorResponse "Failed to mint access token"
// @Router              /api/oauth/token [post]
func HandleToken(c fiber.Ctx) error {
	var body TokenExchangeRequest
	if err := bindOAuthForm(c, &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	if body.GrantType != "" && body.GrantType != "authorization_code" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "unsupported_grant_type"})
	}

	clientID, clientSecret := clientCredsFromRequest(c, body.ClientID, body.ClientSecret)
	if body.Code == "" || clientID == "" || clientSecret == "" || body.RedirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	valid, err := db.VerifyClientCredentials(context.Background(), clientID, clientSecret)
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
	if payload.ClientID != clientID || payload.RedirectURI != body.RedirectURI {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_grant"})
	}

	accessToken, err := mintAccessJWT(payload.Email, clientID, payload.Scopes, accessTokenTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "token_mint_failed"})
	}

	response := TokenExchangeResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   accessTokenTTLSeconds,
		Scope:       joinScopes(payload.Scopes),
	}

	if hasScope(payload.Scopes, ScopeOpenID) {
		profile, err := db.GetRootProfileByEmail(context.Background(), payload.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "profile_lookup_failed"})
		}
		idToken, err := mintIDToken(profile, clientID, payload.Scopes, accessTokenTTL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "token_mint_failed"})
		}
		response.IDToken = idToken
	}

	return c.JSON(response)
}

func bindOAuthForm(c fiber.Ctx, out any) error {
	ct := c.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/json"),
		strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		return c.Bind().Body(out)
	default:
		return errUnsupportedContentType
	}
}

func clientCredsFromRequest(c fiber.Ctx, bodyID, bodySecret string) (string, string) {
	if h := c.Get("Authorization"); strings.HasPrefix(h, "Basic ") {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.TrimPrefix(h, "Basic ")))
		if err == nil {
			if id, secret, ok := strings.Cut(string(raw), ":"); ok && id != "" {
				return id, secret
			}
		}
	}
	return bodyID, bodySecret
}

func joinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

func resolveOAuthScopes(scopeParam string) ([]string, bool) {
	if scopeParam == "" {
		return append([]string(nil), defaultOAuthScopes...), true
	}

	scopes := normalizeScopes(strings.Fields(scopeParam))
	if !validateOAuthScopes(scopes) {
		return nil, false
	}
	return scopes, true
}
