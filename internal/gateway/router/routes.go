package router

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/markosoft2000/bike-tracker/internal/config"
	auth_handler "github.com/markosoft2000/bike-tracker/internal/gateway/handler/auth"
	"github.com/markosoft2000/bike-tracker/internal/gateway/middleware"
)

func SetupRoutes(
	cfg *config.Config,
	log *slog.Logger,
	router fiber.Router,
	auth auth_handler.AuthHandlerService,
) {

	router.Get("/health", func(c fiber.Ctx) error { return c.SendStatus(200) })

	// router.Get("/ip-check", func(c fiber.Ctx) error {
	// 	return c.JSON(fiber.Map{
	// 		"c.IP()":          c.IP(),                             // Should show real client IP
	// 		"X-Forwarded-For": c.Get("X-Forwarded-For"),           // Raw header value
	// 		"IsProxyTrusted":  c.IsProxyTrusted(),                 // Should be true
	// 		"RemoteIP":        c.RequestCtx().RemoteIP().String(), // Proxy IP
	// 	})
	// })

	authGroup := router.Group("/api/v1", middleware.RateLimiter(&cfg.Middleware))
	{
		// Public Authentication Endpoints
		authGroup.Post("/register", auth.Register)
		authGroup.Post("/login", auth.Login)
		authGroup.Post("/refresh", auth.Refresh)

		// Protected Authentication Endpoints (Require valid Bearer token)
		protectedAuth := authGroup.Use(middleware.AuthGuard(log))

		protectedAuth.Post("/logout", auth.Logout)
		protectedAuth.Get("/users/:userId/admin", auth.IsAdmin)

		// Administrative Application Management
		protectedAuth.Post("/apps", auth.AddApp)
		protectedAuth.Delete("/apps/:id", auth.RemoveApp)
	}
}
