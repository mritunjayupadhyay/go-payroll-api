// Package logger builds the application's structured logger.
//
// Output is JSON on stdout so container runtimes and log aggregators can
// parse it without extra config. Level is read from the LOG_LEVEL env var
// (debug, info, warn, error) and defaults to info.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON *slog.Logger configured from the environment.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelFromEnv(),
	}))
}

func levelFromEnv() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
