package api

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/helpers/environment"
)

// RegisterWellKnownRoutes registers OIDC discovery and JWKS endpoints.
func RegisterWellKnownRoutes(app *fiber.App) {
	app.Get("/.well-known/openid-configuration", HandleOpenIDConfiguration)
	app.Get("/.well-known/jwks.json", HandleJWKS)
}

// HandleOpenIDConfiguration returns the OIDC provider metadata document.
//
// @Summary             OIDC Discovery
// @Description         Returns the OpenID Connect provider configuration document.
// @Tags                OIDC
// @Produce             json
// @Success             200 {object} map[string]any
// @Router              /.well-known/openid-configuration [get]
func HandleOpenIDConfiguration(c fiber.Ctx) error {
	issuer := environment.IssuerURL
	return c.JSON(map[string]any{
		"issuer":                                issuer,
		"authorization_endpoint":                fmt.Sprintf("%s/api/oauth/authorize", issuer),
		"token_endpoint":                        fmt.Sprintf("%s/api/oauth/token", issuer),
		"userinfo_endpoint":                     fmt.Sprintf("%s/api/oauth/userinfo", issuer),
		"jwks_uri":                              fmt.Sprintf("%s/.well-known/jwks.json", issuer),
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"scopes_supported":                      []string{"openid", "profile", "email", "read"},
		"claims_supported":                      []string{"sub", "email", "email_verified", "name", "picture", "preferred_username"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"subject_types_supported":               []string{"public"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
	})
}

// HandleJWKS returns the JSON Web Key Set for verifying RS256 tokens.
//
// @Summary             JWKS
// @Description         Returns the public signing keys used by this OIDC provider.
// @Tags                OIDC
// @Produce             json
// @Success             200 {object} map[string]any
// @Failure             500 {object} ErrorResponse
// @Router              /.well-known/jwks.json [get]
func HandleJWKS(c fiber.Ctx) error {
	jwk, err := publicJWK()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "jwks_unavailable"})
	}
	return c.JSON(map[string]any{"keys": []map[string]any{jwk}})
}
