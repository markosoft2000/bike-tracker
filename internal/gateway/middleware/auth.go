package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	libjwt "github.com/markosoft2000/bike-tracker/internal/lib/jwt"
)

// AuthGuard halts unauthorized requests immediately before they hit proxy handlers.
func AuthGuard(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {

		const op = "auth.RegisterNewUser"
		log = log.With(slog.String("op", op))

		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.
				Status(fiber.StatusUnauthorized).
				SendString("Unauthorized: Missing or invalid token format")
		}

		tokenStr := authHeader[7:]

		token, claims, err := libjwt.ParseToken(tokenStr)
		if err != nil || !token.Valid {
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

		if err := validateToken(claims); err != nil {
			return c.
				Status(fiber.StatusUnauthorized).
				JSON(
					fiber.Map{
						"error": fmt.Sprintf("Unauthorized: %v", err),
					},
				)
		}

		c.Locals("userId", claims.UserID) // TODO use userId in handlers for logout ...
		c.Locals("appId", claims.AppID.String())

		return c.Next()
	}
}

func validateToken(claims *libjwt.CustomTokenClaims) error {
	if claims == nil {
		return errors.New("claims are empty")
	}

	v := jwt.NewValidator(
		jwt.WithIssuer("markosoft2000"),
		jwt.WithAudience("auth-service"),
		jwt.WithExpirationRequired(),
	)
	if err := v.Validate(claims); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if claims.UserID == uuid.Nil || claims.AppID == uuid.Nil || claims.Email == "" {
		return errors.New("token is missing mandatory custom claims (sub, app_id, or email)")
	}

	return nil
}
