package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/markosoft2000/bike-tracker/internal/config"
	libjwt "github.com/markosoft2000/bike-tracker/internal/lib/jwt"
	"github.com/markosoft2000/bike-tracker/internal/storage"
)

// AuthGuard halts unauthorized requests immediately before they hit proxy handlers.
func AuthGuard(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.MiddlewareConfig,
	store storage.AppPublicKeyStorage,
) fiber.Handler {
	const op = "gateway.middleware.AuthGuard"

	log = log.With(slog.String("op", op))

	getPK := func(appID string) ([]byte, error) {
		return store.AppPublicKey(ctx, appID)
	}

	return func(c fiber.Ctx) error {

		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.
				Status(fiber.StatusUnauthorized).
				SendString("Unauthorized: Missing or invalid token format")
		}

		tokenStr := authHeader[7:]

		token, claims, err := libjwt.ParseToken(tokenStr, getPK)
		if err != nil {
			log.Error("failed to parse token", slog.Any("error", err))

			return c.
				Status(fiber.StatusUnauthorized).
				JSON(
					fiber.Map{
						"error": "Unauthorized: Failed to parse access token",
					},
				)
		}

		if !token.Valid {
			log.Error("invalid token", slog.Any("token", tokenStr))

			return c.
				Status(fiber.StatusUnauthorized).
				JSON(
					fiber.Map{
						"error": "Unauthorized: Invalid or expired access token",
					},
				)
		}

		log.Debug("token", slog.Any("claims", claims))

		if err := validateToken(cfg, claims); err != nil {
			return c.
				Status(fiber.StatusUnauthorized).
				JSON(
					fiber.Map{
						"error": fmt.Sprintf("Unauthorized: %v", err),
					},
				)
		}

		c.Locals("userId", claims.UserID)
		c.Locals("appId", claims.AppID.String())

		return c.Next()
	}
}

func validateToken(
	cfg *config.MiddlewareConfig,
	claims *libjwt.CustomTokenClaims,
) error {
	if claims == nil {
		return errors.New("claims are empty")
	}

	v := jwt.NewValidator(
		jwt.WithIssuer(cfg.TokenIssuer),
		jwt.WithAudience(cfg.TokenAudience),
		jwt.WithExpirationRequired(),
	)
	if err := v.Validate(claims); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if claims.UserID == uuid.Nil || claims.AppID == uuid.Nil || claims.Email == "" {
		return errors.New("token is missing mandatory custom claims (sub, app_id, or email)")
	}

	// @TODO check if the token is on the blacklist

	return nil
}
