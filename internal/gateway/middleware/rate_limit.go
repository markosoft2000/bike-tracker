package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/markosoft2000/bike-tracker/internal/config"
)

// RateLimiter tracks user IP addresses in a memory-efficient map window.
func RateLimiter(cfg *config.MiddlewareConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		// Allow maximum requests per period of time for unique client IP
		Max:        cfg.RateLimitMax,
		Expiration: cfg.RateLimitExpiration,

		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},

		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).SendString("Too Many Requests: Rate limit exceeded")
		},

		// Skip rate limiting entirely if the client is our local load testing tool
		Next: func(c fiber.Ctx) bool {
			return string(c.Request().Header.UserAgent()) == "fasthttp"
		},
	})
}
