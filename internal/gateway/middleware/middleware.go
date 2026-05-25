package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func SetupMiddleware(srv fiber.Router) {
	srv.Use(recover.New())

	srv.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "DELETE", "PUT", "PATCH", "HEAD", "OPTIONS"},
		MaxAge:       86400, // Caches CORS headers for 24 hours on the client side
	}))

	srv.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",

		// OPTIONAL SPEED TRICK: Mute the logger during bombardier load tests to save CPU
		Next: func(c fiber.Ctx) bool {
			// Skip logging if the client is our bombardier benchmark tool
			return string(c.Request().Header.UserAgent()) == "fasthttp"
		},
	}))
}
