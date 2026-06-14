package api

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/db"
)

// RegisterClientRoutes registers OAuth client management endpoints.
func RegisterClientRoutes(app *fiber.App) {
	clients := app.Group("/api/clients")
	clients.Get("/:id", GetClientInfo)
	clients.Post("/", RequireSession, CreateClient)
}

// GetClientInfo returns the public-facing name of an OAuth client.
//
// @Summary             Get Client Info
// @Description         Returns the display name of an OAuth client by ID. No authentication required.
// @Tags                Clients
// @Produce             json
// @Param               id path string true "Client ID"
// @Success             200 {object} map[string]string
// @Failure             404 {object} ErrorResponse "Client not found"
// @Router              /api/clients/{id} [get]
func GetClientInfo(c fiber.Ctx) error {
	clientID := c.Params("id")
	client, err := db.GetClientByID(context.Background(), clientID)
	if err != nil {
		if errors.Is(err, db.ErrClientNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "client_not_found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_lookup_failed"})
	}
	return c.JSON(fiber.Map{"name": client.Name})
}

// CreateClient creates a new OAuth client owned by the authenticated account.
//
// @Summary             Create OAuth Client
// @Description         Creates a new OAuth client owned by the authenticated account. Requires the clients scope.
// @Tags                Clients
// @Accept              json
// @Produce             json
// @Param               Authorization header string true "Bearer <token> where token is a Session.token"
// @Param               body body CreateClientRequest true "Client name and registered redirect URIs"
// @Success             200 {object} CreateClientResponse "Returns the generated client credentials"
// @Failure             400 {object} ErrorResponse "Missing or invalid client name or redirect URIs"
// @Failure             401 {object} ErrorResponse "Unauthorized due to missing or invalid session"
// @Failure             403 {object} ErrorResponse "Session lacks the clients scope"
// @Failure             500 {object} ErrorResponse "Failed to create client in the graph"
// @Router              /api/clients [post]
func CreateClient(c fiber.Ctx) error {
	sessionToken := c.Locals("sessionToken").(string)
	email := c.Locals("email").(string)

	hasClients, err := db.SessionHasScope(context.Background(), sessionToken, ScopeClients)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "session_lookup_failed"})
	}
	if !hasClients {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "missing_clients_scope"})
	}

	var body CreateClientRequest
	if err := c.Bind().Body(&body); err != nil || body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	redirectURIs := normalizeRedirectURIs(body.RedirectURIs)
	if len(redirectURIs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid_request"})
	}

	clientID := generateSecureUUID()
	clientSecret := generateSecureUUID() + generateSecureUUID()

	client := db.Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         body.Name,
		RedirectURIs: redirectURIs,
	}

	if err := db.CreateClientForRoot(context.Background(), email, client); err != nil {
		if errors.Is(err, db.ErrRootNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_create_failed"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "client_create_failed"})
	}

	return c.JSON(CreateClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         body.Name,
		RedirectURIs: redirectURIs,
	})
}

func normalizeRedirectURIs(uris []string) []string {
	seen := make(map[string]struct{}, len(uris))
	out := make([]string, 0, len(uris))
	for _, uri := range uris {
		uri = strings.TrimSpace(uri)
		if uri == "" {
			continue
		}
		if _, ok := seen[uri]; ok {
			continue
		}
		seen[uri] = struct{}{}
		out = append(out, uri)
	}
	return out
}
