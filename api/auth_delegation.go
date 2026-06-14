package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/db"
	"github.com/maslotwi/graph-auth/helpers/environment"
)

const (
	ScopeRead    = "read"
	ScopeFertile = "fertile"
	ScopeClients = "clients"
)

var allowedScopes = map[string]struct{}{
	ScopeRead:    {},
	ScopeFertile: {},
	ScopeClients: {},
}

// RegisterDelegationRoutes adds your new camera-less SSO endpoints
func RegisterDelegationRoutes(app *fiber.App) {
	auth := app.Group("/api/auth/session")

	auth.Post("/generate-code", RequireSession, GenerateDelegationCode) // Auth required
	auth.Post("/consume-code", ConsumeDelegationCode)                   // Public
}

// GenerateDelegationCode handles Scenario 1 & 2 (Step 1): Active device creates an invitation code
// @Summary             Generate Session Invitation Code
// @Description         Generates a temporary, single-use 6-digit code or URL link from an active, fertile session to invite a new device. Requested scopes are chosen when the code is consumed.
// @Tags                Session Delegation
// @Produce             json
// @Param               Authorization header string true "Bearer <token> where token is a Session.token"
// @Success             200 {object} GenerateDelegationCodeResponse "Returns the 6-digit code, direct link, and TTL expiry"
// @Failure             401 {object} ErrorResponse "Unauthorized due to missing, invalid, or inactive session"
// @Failure             403 {object} ErrorResponse "Parent session lacks the fertile scope required to delegate"
// @Failure             500 {object} ErrorResponse "Internal failure generating crypto code or saving to cache"
// @Router              /api/auth/session/generate-code [post]
func GenerateDelegationCode(c fiber.Ctx) error {
	parentSessionToken := c.Locals("sessionToken").(string)
	if !verifyTokenInGraph(parentSessionToken) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session_invalid"})
	}

	hasFertile, err := db.SessionHasScope(context.Background(), parentSessionToken, ScopeFertile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_lookup_failed"})
	}
	if !hasFertile {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "parent_not_fertile"})
	}

	sixDigitCode, err := generateSixDigitCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "crypto_failure"})
	}

	payload, err := json.Marshal(delegationPayload{
		Parent: parentSessionToken,
		Email:  c.Locals("email").(string),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "serialization_failure"})
	}

	err = storeInRedis("delegate:"+sixDigitCode, string(payload), 120)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "cache_failure"})
	}

	return c.JSON(GenerateDelegationCodeResponse{
		Code:      sixDigitCode,
		Link:      fmt.Sprintf("%s/join?code=%s", environment.FrontendURL, sixDigitCode),
		ExpiresIn: 120,
	})
}

// ConsumeDelegationCode handles Scenario 1 & 2 (Step 2): New device redeems the invitation code
// @Summary             Consume Session Invitation Code
// @Description         Redeems a valid 6-digit delegation code to provision a new child session node in the Neo4j provenance graph. Requested scopes must be limited to read and fertile and be a subset of the parent session scopes.
// @Tags                Session Delegation
// @Accept              json
// @Produce             json
// @Param               body body ConsumeDelegationCodeRequest true "JSON body containing the 6-digit code, device name, and requested scopes"
// @Success             200 {object} ConsumeDelegationCodeResponse "Returns a brand new persistent session token and active scopes"
// @Failure             400 {object} ErrorResponse "Invalid request format or requested scopes are not allowed"
// @Failure             401 {object} ErrorResponse "The code has expired, been used, or is mathematically invalid"
// @Failure             403 {object} ErrorResponse "Parent session was revoked, inactive, lacks fertile scope, or requested scopes exceed parent scopes"
// @Router              /api/auth/session/consume-code [post]
func ConsumeDelegationCode(c fiber.Ctx) error {
	var body ConsumeDelegationCodeRequest
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	if body.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	payload, err := consumeFromRedis("delegate:" + body.Code)
	if err != nil || payload == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "code_expired_or_invalid"})
	}

	var delegation delegationPayload
	if err := json.Unmarshal([]byte(payload), &delegation); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "code_expired_or_invalid"})
	}

	deviceName := body.DeviceName
	if deviceName == "" {
		deviceName = "New Device"
	}

	requestedScopes := normalizeScopes(body.Scopes)
	if !validateScopes(requestedScopes) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_scope"})
	}

	parentScopes, active, err := db.ActiveSessionScopes(context.Background(), delegation.Parent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_lookup_failed"})
	}
	if !active {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "session_delegation_denied_or_parent_revoked"})
	}
	if !isSubsetScopes(requestedScopes, parentScopes) {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "scope_not_inherited"})
	}

	newDeviceSessionToken := generateSecureUUID()
	childSession := db.Session{
		Token:      newDeviceSessionToken,
		DeviceName: deviceName,
		Scopes:     requestedScopes,
		IsActive:   true,
	}
	err = db.CreateChildSession(context.Background(), delegation.Parent, childSession)
	if err != nil {
		if errors.Is(err, db.ErrParentNotFertile) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "parent_not_fertile"})
		}
		if errors.Is(err, db.ErrChildScopesNotSubset) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "scope_not_inherited"})
		}
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "session_delegation_denied_or_parent_revoked"})
	}

	if delegation.Email != "" {
		if err := storeInRedis("session:"+newDeviceSessionToken, delegation.Email, sessionCacheTTLSeconds); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_store_failed"})
		}
	} else if email, active, err := db.ActiveSessionEmail(context.Background(), newDeviceSessionToken); err == nil && active && email != "" {
		if err := storeInRedis("session:"+newDeviceSessionToken, email, sessionCacheTTLSeconds); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_store_failed"})
		}
	}

	return c.JSON(ConsumeDelegationCodeResponse{
		SessionToken: newDeviceSessionToken,
		Scopes:       requestedScopes,
		Status:       "authenticated",
	})
}

func normalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}

func validateScopes(scopes []string) bool {
	for _, scope := range scopes {
		if _, ok := allowedScopes[scope]; !ok {
			return false
		}
	}
	return true
}

func isSubsetScopes(child, parent []string) bool {
	parentSet := make(map[string]struct{}, len(parent))
	for _, scope := range parent {
		parentSet[scope] = struct{}{}
	}
	for _, scope := range child {
		if _, ok := parentSet[scope]; !ok {
			return false
		}
	}
	return true
}

func hasScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func withScope(scopes []string, scope string) []string {
	if hasScope(scopes, scope) {
		return scopes
	}
	return append(scopes, scope)
}

func generateSixDigitCode() (string, error) {
	maxNum := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, maxNum)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", n.Int64()+100000), nil
}
