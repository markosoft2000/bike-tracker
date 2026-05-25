package logger

import (
	"io"
	"log/slog"
	"os"
)

const (
	envProd = "prod"
)

// Setup initializes a global structured log handler based on the environment string.
func Setup(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		// Human-readable structured text layout for local development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)

	// Set as global default logger across standard libraries
	slog.SetDefault(logger)

	return logger
}

// NewDiscardLogger creates a completely silent logger wrapper.
// Highly useful for absolute zero-allocation benchmark scenarios.
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
