package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/maslotwi/graph-auth/db"
	"github.com/redis/go-redis/v9"
)

// RegisterOAuthRoutes registers the central SSO engine routes
func RegisterOAuthRoutes(app *fiber.App) {
	oauth := app.Group("/api/oauth")

	oauth.Get("/authorize", HandleAuthorize)
	oauth.Post("/confirm", HandleConfirmLogin)
	oauth.Post("/token", HandleTokenExchange)
}

// HandleAuthorize handles the initial incoming OIDC/OAuth2 redirect from external apps.
// @Summary             Authorize SSO Request
// @Description         Evaluates device state. Redirects authenticated devices to a consent screen, and unauthenticated devices to the session delegation flow.
// @Tags                OAuth2 SSO
// @Produce             json
// @Param               client_id query string true "OAuth2 Client ID"
// @Param               redirect_uri query string true "OAuth2 Redirect URI"
// @Param               state query string true "OAuth2 State string"
// @Param               X-Session-Token header string false "Existing Neo4j Session Token"
// @Success             200 {object} AuthorizeResponse "Returns the status and the next frontend URL to redirect to"
// @Failure             400 {object} ErrorResponse "Missing required OAuth2 parameters"
// @Router              /api/oauth/authorize [get]
func HandleAuthorize(c fiber.Ctx) error {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	state := c.Query("state")

	if clientID == "" || redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_oauth_parameters"})
	}

	// Check if this specific browser already holds a valid root session token
	// In production, this token will be fetched from an HTTP-Only cookie or header
	existingDeviceToken := c.Get("X-Session-Token")

	// Check Neo4j to see if this token is alive and unrevoked
	isDeviceAuthenticated := false
	if existingDeviceToken != "" {
		isDeviceAuthenticated = verifyTokenInGraph(existingDeviceToken)
	}

	// Construct the query tracking parameters so the frontend knows where to return when done
	flowContext := fmt.Sprintf("?client_id=%s&redirect_uri=%s&state=%s", clientID, redirectURI, state)

	// CASE 1: Device is already verified in your Graph
	if isDeviceAuthenticated {
		return c.JSON(AuthorizeResponse{
			Status:  "authenticated",
			Message: "Device is recognized. Prompt user for consent.",
			NextURL: "/frontend/sso-consent" + flowContext,
		})
	}

	return c.JSON(AuthorizeResponse{
		Status:  "unauthenticated",
		Message: "Device unrecognized. Pair this device with an active device to log in.",
		NextURL: "/frontend/link-device" + flowContext,
	})
}

// HandleConfirmLogin is triggered when the user clicks the "Confirm Login" button on the consent page.
// @Summary             Confirm SSO Login
// @Description         Verifies the active session and generates a temporary, high-entropy OAuth2 authorization code.
// @Tags                OAuth2 SSO
// @Accept              json
// @Produce             json
// @Param               X-Session-Token header string true "Active Neo4j Session Token"
// @Param               body body ConfirmLoginRequest true "JSON body containing client_id, redirect_uri, and state"
// @Success             200 {object} ConfirmLoginResponse "Returns the callback URL containing the auth code"
// @Failure             400 {object} ErrorResponse "Invalid payload"
// @Failure             401 {object} ErrorResponse "Session invalid or revoked"
// @Router              /api/oauth/confirm [post]
func HandleConfirmLogin(c fiber.Ctx) error {
	var body ConfirmLoginRequest
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_payload"})
	}

	deviceToken := c.Get("X-Session-Token")
	if !verifyTokenInGraph(deviceToken) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session_invalidated_during_flow"})
	}

	// 1. Generate a temporary, high-entropy OAuth Authorization Code
	authCode := "code_" + generateSecureUUID()

	// 2. Map this authorization code to the specific User ID and Client App inside Redis with a 30-second TTL
	// In production, fetch the actual User ID linked to this deviceToken from your Neo4j path
	mockUserID := "user_999"
	redisPayload := fmt.Sprintf(`{"user_id":"%s","client_id":"%s"}`, mockUserID, body.ClientID)
	_ = storeInRedis("code:"+authCode, redisPayload, 30)

	// 3. Hand back the callback target. The frontend will physically redirect the browser here.
	callbackURL := fmt.Sprintf("%s?code=%s&state=%s", body.RedirectURI, authCode, body.State)
	return c.JSON(ConfirmLoginResponse{
		Status:     "success",
		RedirectTo: callbackURL,
	})
}

// HandleTokenExchange is called server-to-server by Service A backend to trade the code for an Access JWT.
// @Summary             Exchange Auth Code for JWT
// @Description         Consumes a short-lived authorization code and issues a signed JWT access token backed by Redis.
// @Tags                OAuth2 SSO
// @Accept              json
// @Produce             json
// @Param               body body TokenExchangeRequest true "JSON body containing code, client_id, and client_secret"
// @Success             200 {object} TokenExchangeResponse "Returns the standard OAuth2 JWT access token payload"
// @Failure             400 {object} ErrorResponse "Invalid request format"
// @Failure             401 {object} ErrorResponse "Invalid, expired, or already consumed authorization code"
// @Router              /api/oauth/token [post]
func HandleTokenExchange(c fiber.Ctx) error {
	var body TokenExchangeRequest
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	// 1. Fetch the code mapping out of Redis
	payload, err := getFromRedis("code:" + body.Code)
	if err != nil || payload == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "invalid_or_expired_code"})
	}

	// 2. Immediately burn the code so it cannot be replayed (Strict OAuth2 Security)
	_ = deleteFromRedis("code:" + body.Code)

	// 3. Extract metadata from payload, run your cryptography minting logic, and register the short-lived session in Redis
	tempSessionID := "access_session_" + generateSecureUUID()
	mockUserID := "user_999"
	mockScopes := []string{"read"}

	jwtToken, _ := mintJWT(mockUserID, tempSessionID, mockScopes)
	_ = storeInRedis("access:"+tempSessionID, "valid", 900) // 15 Minute session

	// 4. Return standard compliant OAuth2 access token details back to Service A
	return c.JSON(TokenExchangeResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   900,
		Scopes:      mockScopes,
	})
}

// ------------------------------------------------------------------------
// SESSION CACHE AND GRAPH VERIFICATION
// ------------------------------------------------------------------------

const sessionCacheTTLSeconds = 86400

func verifyTokenInGraph(token string) bool {
	if token == "" {
		return false
	}

	_, active, err := db.ActiveSessionEmail(context.Background(), token)
	return err == nil && active
}

func resolveSessionEmail(token string) (string, bool) {
	email, err := getFromRedis("session:" + token)
	if err == nil && email != "" {
		return email, true
	}

	email, active, err := db.ActiveSessionEmail(context.Background(), token)
	if err != nil || !active || email == "" {
		return "", false
	}

	_ = storeInRedis("session:"+token, email, sessionCacheTTLSeconds)
	return email, true
}

func generateSecureUUID() string {
	return uuid.NewString()
}

func storeInRedis(key string, value string, ttlSeconds int) error {
	client, err := db.Redis()
	if err != nil {
		return err
	}
	return client.Set(context.Background(), key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

func getFromRedis(key string) (string, error) {
	client, err := db.Redis()
	if err != nil {
		return "", err
	}
	val, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func deleteFromRedis(key string) error {
	client, err := db.Redis()
	if err != nil {
		return err
	}
	return client.Del(context.Background(), key).Err()
}

func consumeFromRedis(key string) (string, error) {
	client, err := db.Redis()
	if err != nil {
		return "", err
	}
	val, err := client.GetDel(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func extractSessionFromJWT(authHeader string) (string, error) {
	// TODO: Parse JWT, verify signature, return 'jti' claim
	return "phone_session_123", nil
}

func createDelegatedSessionInGraph(parentSessionID string, deviceName string, scopes []string) (string, error) {
	// TODO: Cypher -> MATCH parent CREATE child CREATE (parent)-[:SPAWNED]->(child)
	return "new_pc_refresh_token_789", nil
}

func validateGraphLineage(refreshToken string) (string, []string, error) {
	// TODO: Cypher -> Traverse graph upwards to Root, ensure no session is deactivated
	return "user_123", []string{"read"}, nil
}

func mintJWT(userID string, sessionID string, scopes []string) (string, error) {
	// TODO: Use golang-jwt/jwt/v5 to sign token with RSA private key
	return "eyJhbGciOiJSUzI1...", nil
}
