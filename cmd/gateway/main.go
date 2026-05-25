package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	appMain "github.com/markosoft2000/bike-tracker/internal/app"
	"github.com/markosoft2000/bike-tracker/internal/config"
	"github.com/markosoft2000/bike-tracker/internal/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	log := logger.Setup(cfg.Env)

	app := appMain.New(ctx, log, cfg)
	app.MustRun()

	// Graceful Shutdown Setup
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Wait for SIGINT or SIGTERM
	<-sigCtx.Done()

	app.Stop(ctx)
}
