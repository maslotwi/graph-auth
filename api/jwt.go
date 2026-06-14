package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maslotwi/graph-auth/db"
	"github.com/maslotwi/graph-auth/helpers/environment"
)

var (
	errJWTInvalid    = errors.New("invalid jwt")
	errJWTRevoked    = errors.New("jwt revoked")
	errJWTMissingJTI = errors.New("jwt missing jti")
)

// AccessClaims are the custom claims embedded in an OAuth access JWT.
type AccessClaims struct {
	Email    string   `json:"email"`
	ClientID string   `json:"client_id"`
	Scopes   []string `json:"scopes"`
	jwt.RegisteredClaims
}

// IDClaims are the standard OIDC claims embedded in an id_token.
type IDClaims struct {
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	Picture           string `json:"picture,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	jwt.RegisteredClaims
}

func mintAccessJWT(email, clientID string, scopes []string, ttl time.Duration) (string, error) {
	key, err := signingKey()
	if err != nil {
		return "", err
	}
	kid, err := keyID()
	if err != nil {
		return "", err
	}

	jti := generateSecureUUID()
	now := time.Now()

	claims := AccessClaims{
		Email:    email,
		ClientID: clientID,
		Scopes:   scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			Audience:  jwt.ClaimStrings{clientID},
			Issuer:    environment.IssuerURL,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	ttlSeconds := int(ttl.Seconds())
	if ttlSeconds <= 0 {
		ttlSeconds = 1
	}
	if err := storeInRedis("jwt:"+jti, "valid", ttlSeconds); err != nil {
		return "", err
	}

	return signed, nil
}

func mintIDToken(profile db.RootProfile, clientID string, scopes []string, ttl time.Duration) (string, error) {
	key, err := signingKey()
	if err != nil {
		return "", err
	}
	kid, err := keyID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := IDClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   profile.Email,
			Audience:  jwt.ClaimStrings{clientID},
			Issuer:    environment.IssuerURL,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	if hasScope(scopes, ScopeEmail) {
		claims.Email = profile.Email
		claims.EmailVerified = true
	}
	if hasScope(scopes, ScopeProfile) {
		claims.Name = profile.DisplayName
		claims.Picture = profile.Picture
		claims.PreferredUsername = preferredUsernameFromEmail(profile.Email)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	return token.SignedString(key)
}

func parseAccessJWT(tokenString string) (*AccessClaims, error) {
	key, err := signingKey()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &key.PublicKey, nil
	})
	if err != nil {
		return nil, errJWTInvalid
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errJWTInvalid
	}
	if claims.ID == "" {
		return nil, errJWTMissingJTI
	}

	nonce, err := getFromRedis("jwt:" + claims.ID)
	if err != nil {
		return nil, err
	}
	if nonce == "" {
		return nil, errJWTRevoked
	}

	return claims, nil
}

func revokeAccessJWT(jti string) error {
	if jti == "" {
		return errJWTMissingJTI
	}
	return deleteFromRedis("jwt:" + jti)
}
