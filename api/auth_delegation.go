package api

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/gofiber/fiber/v3"
)

// RegisterDelegationRoutes adds your new camera-less SSO endpoints
func RegisterDelegationRoutes(app *fiber.App) {
	auth := app.Group("/api/auth/session")

	auth.Post("/generate-code", GenerateDelegationCode) // Called by Computer A (Auth Required)
	auth.Post("/consume-code", ConsumeDelegationCode)   // Called by Computer B or Phone B (Public)
}

// GenerateDelegationCode handles Scenario 1 & 2 (Step 1): Active device creates an invitation code
// @Summary             Generate Session Invitation Code
// @Description         Generates a temporary, single-use 6-digit code or URL link from an active session to invite a new device.
// @Tags                Session Delegation
// @Accept              json
// @Produce             json
// @Param               X-Session-Token header string true "Active Session Token of Generator Device"
// @Param               body body object false "Desired scopes for the target device"
// @Success             200 {object} map[string]interface{} "Returns the 6-digit code, direct link, and TTL expiry"
// @Failure             401 {object} map[string]string "Unauthorized due to missing or invalid session context"
// @Failure             500 {object} map[string]string "Internal failure generating crypto code or saving to cache"
// @Router              /api/auth/session/generate-code [post]
func GenerateDelegationCode(c fiber.Ctx) error {
	// 1. Identify who is generating this code via their active session token
	parentSessionToken := c.Get("X-Session-Token")
	if parentSessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing_session_context"})
	}

	var body struct {
		Scopes []string `json:"scopes"` // e.g., ["read", "write"] or restricted ["read"]
	}
	if err := c.Bind().Body(&body); err != nil {
		body.Scopes = []string{"read"} // fallback to safe default
	}

	// 2. Generate a secure, user-friendly 6-digit numeric code
	sixDigitCode, err := generateSixDigitCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "crypto_failure"})
	}

	// 3. Store the delegation intent in Redis linked to Parent's Token
	// In production, serialize this as a JSON string containing parentSessionToken and body.Scopes
	redisPayload := fmt.Sprintf(`{"parent":"%s","scopes":["%s"]}`, parentSessionToken, body.Scopes[0])
	err = storeInRedis("delegate:"+sixDigitCode, redisPayload, 120) // 2 minute expiry
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cache_failure"})
	}

	// 4. Return the code and the absolute URL that Phone B can open directly if rendered as a QR code
	return c.JSON(fiber.Map{
		"code":       sixDigitCode,
		"link":       fmt.Sprintf("https://yourdomain.com/login?code=%s", sixDigitCode),
		"expires_in": 120,
	})
}

// ConsumeDelegationCode handles Scenario 1 & 2 (Step 2): New device redeems the invitation code
// @Summary             Consume Session Invitation Code
// @Description         Redeems a valid 6-digit delegation code to provision a new child session node in the Neo4j provenance graph.
// @Tags                Session Delegation
// @Accept              json
// @Produce             json
// @Param               body body object true "JSON body containing the 6-digit code and identifying device name"
// @Success             200 {object} map[string]interface{} "Returns a brand new persistent session token and active scopes"
// @Failure             401 {object} map[string]string "The code has expired, been used, or is mathematically invalid"
// @Failure             403 {object} map[string]string "Graph constraints prevented attachment (e.g. parent session was revoked)"
// @Router              /api/auth/session/consume-code [post]
func ConsumeDelegationCode(c fiber.Ctx) error {
	var body struct {
		Code       string `json:"code"`
		DeviceName string `json:"device_name"` // e.g., "Computer B" or "Phone B"
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_request"})
	}

	// 1. Fetch metadata from Redis using the code
	payload, err := getFromRedis("delegate:" + body.Code)
	if err != nil || payload == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "code_expired_or_invalid"})
	}

	// 2. Erase code from Redis immediately so it is strictly single-use (One-Time Passcode rule)
	_ = deleteFromRedis("delegate:" + body.Code)

	// 3. Extract the parent session context from the payload (stub logic)
	parentSessionToken := "extracted_from_payload"
	allowedScopes := []string{"read"}

	// 4. Execute the Neo4j insertion logic to append this device directly to the lineage tree
	newDeviceSessionToken := generateSecureUUID()
	err = insertChildSessionIntoGraph(parentSessionToken, newDeviceSessionToken, body.DeviceName, allowedScopes)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "session_delegation_denied_or_parent_revoked"})
	}

	// 5. Hand back a fresh session token to the new device!
	return c.JSON(fiber.Map{
		"session_token": newDeviceSessionToken,
		"scopes":        allowedScopes,
		"status":        "authenticated",
	})
}

// ------------------------------------------------------------------------
// CRYPTO HELPERS
// ------------------------------------------------------------------------

func generateSixDigitCode() (string, error) {
	maxNum := big.NewInt(900000) // 0 to 899999
	n, err := rand.Int(rand.Reader, maxNum)
	if err != nil {
		return "", err
	}
	// Shift up to range 100000 - 999999
	return fmt.Sprintf("%d", n.Int64()+100000), nil
}

// ------------------------------------------------------------------------
// EXTRA STUBS FOR YOUR MAIN CONTEXT
// ------------------------------------------------------------------------

func deleteFromRedis(key string) error {
	// TODO: Implement redis DEL command
	return nil
}

func insertChildSessionIntoGraph(parentToken, childToken, deviceName string, scopes []string) error {
	// TODO: Cypher statement connecting (p:Session {token: parentToken})-[s:SPAWNED]->(c:Session {token: childToken})
	return nil
}
