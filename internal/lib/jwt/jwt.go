package jwt

import (
	"crypto/ed25519"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomTokenClaims struct {
	jwt.RegisteredClaims
	Email  string    `json:"email"`
	UserID uuid.UUID `json:"sub"`
	AppID  uuid.UUID `json:"app_id"`
}

func ParseToken(
	t string,
	getPK func(appID string) ([]byte, error),
) (*jwt.Token, *CustomTokenClaims, error) {
	op := "jwt.ParseToken"

	claims := &CustomTokenClaims{}
	token, err := jwt.ParseWithClaims(t, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		clms, ok := t.Claims.(*CustomTokenClaims)
		if !ok {
			return nil, fmt.Errorf("invalid token claims structure context")
		}

		pk, err := getPK(clms.AppID.String())
		if err != nil {
			return nil, fmt.Errorf("%s: failed to get app public key %w", op, err)
		}

		publicKey, err := getPublicKey(pk)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to parse public key %w", op, err)
		}

		return publicKey, nil
	})

	return token, claims, err
}

func getPublicKey(publicKeyPEM []byte) (any, error) {
	parsedKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to compile gateway cryptographic public trust anchor: %w", err)
	}

	ed25519Pub, ok := parsedKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid public key configuration: key is not of type Ed25519: %w", err)
	}

	return ed25519Pub, nil
}
