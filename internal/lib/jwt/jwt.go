package jwt

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomTokenClaims struct {
	jwt.RegisteredClaims
	Email  string    `json:"email"`
	UserID uuid.UUID `json:"sub"`
	AppID  uuid.UUID `json:"app_id"`
}

func ParseToken(t string) (*jwt.Token, *CustomTokenClaims, error) {
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

		publicKey, err := getPublicKey(clms.AppID)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to get public key %w", op, err)
		}

		return publicKey, nil
	})

	return token, claims, err
}

func getPublicKey(appID uuid.UUID) (any, error) {
	// TODO
	if appID.String() != "019dfd8c-a2ca-7d73-b3c7-80840b1fbed9" {
		return nil, errors.New("invalid app id")
	}

	// TODO get public key for a specific app from redis. update auth keys in redis via kafka
	publicKeyPEM := []byte(`-----BEGIN PUBLIC KEY-----
M...
-----END PUBLIC KEY-----`)

	// Parse once at startup into memory
	parsedKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		log.Fatalf("failed to compile gateway cryptographic public trust anchor: %v", err) // TODO no fatal
	}

	ed25519Pub, ok := parsedKey.(ed25519.PublicKey)
	if !ok {
		log.Fatalf("invalid public key configuration: key is not of type Ed25519") // TODO no fatal
	}

	return ed25519Pub, nil
}
