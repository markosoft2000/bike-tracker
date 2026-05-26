package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/markosoft2000/bike-tracker/internal/config"
	auth_handler "github.com/markosoft2000/bike-tracker/internal/gateway/handler/auth"
	"github.com/markosoft2000/bike-tracker/internal/gateway/middleware"
	"github.com/markosoft2000/bike-tracker/internal/gateway/router"
	libjson "github.com/markosoft2000/bike-tracker/internal/lib/json"
	"github.com/markosoft2000/bike-tracker/internal/storage"
	"github.com/markosoft2000/bike-tracker/internal/storage/redis"
)

type App struct {
	cfg *config.Config
	log *slog.Logger

	httpServer *fiber.App

	// services
	authHandler auth_handler.AuthHandlerService

	redisStorage storage.AppPublicKeyStorage
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
) *App {
	// CONFIG
	srv := fiber.New(fiber.Config{
		ServerHeader:       cfg.HTTPServer.ServerHeader,
		DisableKeepalive:   cfg.HTTPServer.DisableKeepalive,
		Concurrency:        cfg.HTTPServer.Concurrency, // 256 * 1024,
		ReduceMemoryUsage:  cfg.HTTPServer.ReduceMemoryUsage,
		DisableDefaultDate: cfg.HTTPServer.DisableDefaultDate,

		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,

		JSONEncoder: libjson.JSONEncoder,
		JSONDecoder: libjson.JSONDecoder,

		// --- GOOGLE CLOUD LOAD BALANCER COMPATIBILITY ---
		//ProxyHeader: fiber.HeaderXForwardedFor, // Instructs c.IP() to read 'X-Forwarded-For'
	})

	// MIDDLEWARES
	middleware.SetupMiddleware(srv)

	// HANDLERS
	authHandler := auth_handler.NewAuthHandler(log, cfg)

	// Storages
	redis, err := redis.New(redis.Config{
		Host:             cfg.Redis.Host,
		Port:             cfg.Redis.Port,
		OperationTimeout: cfg.Redis.OperationTimeout,
	})
	if err != nil {
		log.Error("failed to init redis", slog.Any("error", err))
		os.Exit(1)
	}

	// TODO update app-pk via kafka
	redis.SaveAppPublicKey(
		ctx,
		"019dfd8c-a2ca-7d73-b3c7-80840b1fbed9",
		[]byte(`-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAnxd77m/ARyaInhO0sCE5sQt3JNeWCSTCU2rtF7lj+X0=
-----END PUBLIC KEY-----`),
	)

	// ROUTER
	router.SetupRoutes(ctx, &cfg.Middleware, log, srv, authHandler, redis)

	return &App{
		cfg: cfg,
		log: log,

		httpServer: srv,

		authHandler: authHandler,

		redisStorage: redis,
	}
}

func (app *App) MustRun() {
	go func() {
		app.log.Info("http server starting", slog.String("addr", app.cfg.HTTPServer.Address))

		listenConfig := fiber.ListenConfig{
			DisableStartupMessage: false,
		}

		if err := app.httpServer.Listen(app.cfg.HTTPServer.Address, listenConfig); err != nil {
			app.log.Error("http server failed", slog.Any("error", err))
		}
	}()
}

func (app *App) Stop(ctx context.Context) {
	app.log.Info("shutting down gracefully...")

	start := time.Now()

	if err := app.httpServer.ShutdownWithContext(ctx); err != nil {
		app.log.Error("forced shutdown http server", "error", err)
	}

	app.authHandler.Close()
	app.redisStorage.Stop()

	app.log.Info("server stopped", slog.Duration("duration", time.Since(start)))
}
