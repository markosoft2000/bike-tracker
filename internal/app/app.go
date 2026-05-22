package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/markosoft2000/bike-tracker/internal/config"
)

type App struct {
	cfg *config.Config
	log *slog.Logger

	httpServer *fiber.App
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
) *App {

	srv := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
	})

	// GET route for the root path
	srv.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	// GET route with a URL parameter
	srv.Get("/user/:name", func(c fiber.Ctx) error {
		name := c.Params("name")
		return c.SendString("Hello, " + name)
	})

	// GET route returning a JSON object
	srv.Get("/json", func(c fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"message": "success",
			"data":    "Fiber is fast!",
		})
	})

	return &App{
		cfg: cfg,
		log: log,

		httpServer: srv,
	}
}

func (app *App) MustRun() {
	go func() {
		app.log.Info("http server starting", slog.String("addr", app.cfg.HTTPServer.Address))

		if err := app.httpServer.Listen(app.cfg.HTTPServer.Address); err != nil {
			app.log.Error("http server failed", slog.Any("error", err))
		}
	}()
}

func (app *App) Stop(ctx context.Context) {
	app.log.Info("shutting down gracefully...")

	start := time.Now()

	if err := app.httpServer.Shutdown(); err != nil {
		app.log.Error("forced shutdown http server", "error", err)
	}

	app.log.Info("server stopped", slog.Duration("duration", time.Since(start)))
}
