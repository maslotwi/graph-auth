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
	Token  string   `json:"token"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// VerifyResponse is returned after a magic-link token is verified.
type VerifyResponse struct {
	SessionToken string `json:"sessionToken"`
	Email        string `json:"email"`
}

// AuthorizeResponse is returned when evaluating an OAuth2 authorize request.
type AuthorizeResponse struct {
	Status  string `json:"status" example:"authenticated"`
	Message string `json:"message"`
	NextURL string `json:"next_url"`
}

// ConfirmLoginRequest is the body for confirming an SSO login.
type ConfirmLoginRequest struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
}

// ConfirmLoginResponse is returned after generating an OAuth2 authorization code.
type ConfirmLoginResponse struct {
	Status     string `json:"status" example:"success"`
	RedirectTo string `json:"redirect_to"`
}

// TokenExchangeRequest is the body for exchanging an authorization code for a JWT.
type TokenExchangeRequest struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// TokenExchangeResponse is the standard OAuth2 access token payload.
type TokenExchangeResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type" example:"Bearer"`
	ExpiresIn   int      `json:"expires_in"`
	Scopes      []string `json:"scopes"`
}

// GenerateDelegationCodeRequest is the optional body for generating a delegation code.
type GenerateDelegationCodeRequest struct {
	Scopes []string `json:"scopes"`
}

// GenerateDelegationCodeResponse is returned when a delegation code is created.
type GenerateDelegationCodeResponse struct {
	Code      string `json:"code"`
	Link      string `json:"link"`
	ExpiresIn int    `json:"expires_in"`
}

// ConsumeDelegationCodeRequest is the body for redeeming a delegation code.
type ConsumeDelegationCodeRequest struct {
	Code       string `json:"code"`
	DeviceName string `json:"device_name"`
}

// ConsumeDelegationCodeResponse is returned when a delegation code is consumed.
type ConsumeDelegationCodeResponse struct {
	SessionToken string   `json:"session_token"`
	Scopes       []string `json:"scopes"`
	Status       string   `json:"status" example:"authenticated"`
}

// delegationPayload is the Redis cache entry for a pending delegation code.
type delegationPayload struct {
	Parent string   `json:"parent"`
	Scopes []string `json:"scopes"`
}
