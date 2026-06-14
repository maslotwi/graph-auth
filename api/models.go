package api

// ErrorResponse is the standard error payload returned by API endpoints.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid_request"`
}

// HealthResponse is returned by the health check endpoint.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// RegisterRequest is the body for requesting a magic-link email.
type RegisterRequest struct {
	Email string `json:"email"`
}

// RegisterResponse is returned after a magic-link email is sent.
type RegisterResponse struct {
	Message string `json:"message" example:"Check your email for a verification link."`
}

// VerifyRequest is the body for exchanging a magic-link token for a session.
type VerifyRequest struct {
	Token       string   `json:"token"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Picture     string   `json:"picture"`
	Scopes      []string `json:"scopes"`
}

// VerifyResponse is returned after a magic-link token is verified.
type VerifyResponse struct {
	SessionToken string `json:"sessionToken"`
	Email        string `json:"email"`
}

// AuthorizeRequest is the body for issuing an OAuth2 authorization code.
type AuthorizeRequest struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
	Scope       string `json:"scope"`
}

// AuthorizeResponse is returned after an OAuth2 authorization code is created.
type AuthorizeResponse struct {
	Status     string `json:"status" example:"success"`
	RedirectTo string `json:"redirect_to"`
}

// TokenExchangeRequest is the body for exchanging an authorization code for a JWT.
type TokenExchangeRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code" form:"code"`
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
}

// TokenExchangeResponse is the standard OAuth2 access token payload.
type TokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token,omitempty"`
}

// GenerateDelegationCodeResponse is returned when a delegation code is created.
type GenerateDelegationCodeResponse struct {
	Code      string `json:"code"`
	Link      string `json:"link"`
	ExpiresIn int    `json:"expires_in"`
}

// ConsumeDelegationCodeRequest is the body for redeeming a delegation code.
type ConsumeDelegationCodeRequest struct {
	Code       string   `json:"code"`
	DeviceName string   `json:"device_name"`
	Scopes     []string `json:"scopes"`
}

// ConsumeDelegationCodeResponse is returned when a delegation code is consumed.
type ConsumeDelegationCodeResponse struct {
	SessionToken string   `json:"session_token"`
	Scopes       []string `json:"scopes"`
	Status       string   `json:"status" example:"authenticated"`
}

// CreateClientRequest is the body for creating an OAuth client.
type CreateClientRequest struct {
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
}

// CreateClientResponse is returned after an OAuth client is created.
type CreateClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
}

// UserInfoResponse is the OIDC userinfo payload returned for access tokens.
type UserInfoResponse struct {
	Sub               string `json:"sub"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	Picture           string `json:"picture,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

// delegationPayload is the Redis cache entry for a pending delegation code.
type delegationPayload struct {
	Parent string `json:"parent"`
	Email  string `json:"email"`
}

// authCodePayload is the Redis cache entry for a pending OAuth authorization code.
type authCodePayload struct {
	Email       string   `json:"email"`
	ClientID    string   `json:"client_id"`
	RedirectURI string   `json:"redirect_uri"`
	Scopes      []string `json:"scopes"`
}
