package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/maslotwi/graph-auth/helpers/environment"
)

var (
	signingKeyOnce sync.Once
	signingKeyErr  error
	privateKey     *rsa.PrivateKey
	publicKeyID    string
)

func initSigningKey() {
	signingKeyOnce.Do(func() {
		privateKey, signingKeyErr = loadOrGenerateSigningKey(environment.RSAPrivateKeyPath)
		if signingKeyErr != nil {
			return
		}
		publicKeyID, signingKeyErr = computeKeyID(&privateKey.PublicKey)
	})
}

func signingKey() (*rsa.PrivateKey, error) {
	initSigningKey()
	return privateKey, signingKeyErr
}

func keyID() (string, error) {
	initSigningKey()
	return publicKeyID, signingKeyErr
}

func publicJWK() (map[string]any, error) {
	initSigningKey()
	if signingKeyErr != nil {
		return nil, signingKeyErr
	}

	kid, err := keyID()
	if err != nil {
		return nil, err
	}

	pub := privateKey.PublicKey
	return map[string]any{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": kid,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}, nil
}

func loadOrGenerateSigningKey(path string) (*rsa.PrivateKey, error) {
	if data, err := os.ReadFile(path); err == nil {
		return parseRSAPrivateKey(data)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	encoded := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return nil, fmt.Errorf("write rsa private key: %w", err)
	}

	return key, nil
}

func parseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid pem block")
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("pem is not an rsa private key")
	}

	return key, nil
}

func computeKeyID(pub *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}
