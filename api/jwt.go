package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func mintAccessJWT(email, clientID string, scopes []string, ttl time.Duration) (string, error) {
	jti := generateSecureUUID()
	now := time.Now()

	claims := AccessClaims{
		Email:    email,
		ClientID: clientID,
		Scopes:   scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(environment.JWTSecret))
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

func parseAccessJWT(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(environment.JWTSecret), nil
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
