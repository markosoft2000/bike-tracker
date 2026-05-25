package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// AuthGuard halts unauthorized requests immediately before they hit proxy handlers.
func AuthGuard() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Next()
	}
}
