package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/handlers"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/logger"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/storage"
)

func main() {
	log := logger.New()
	slog.SetDefault(log)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Error("startup: DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()
	store, err := storage.New(ctx, dsn)
	if err != nil {
		log.Error("startup: connect storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	addr := ":8080"
	log.Info("server starting", "addr", addr)

	r := handlers.NewRouter(handlers.New(store, log))
	if err := r.Run(addr); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
