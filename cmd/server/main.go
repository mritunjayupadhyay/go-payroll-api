package main

import (
	"context"
	"log"
	"os"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/handlers"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/storage"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	store, err := storage.New(ctx, dsn)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	defer store.Close()

	r := handlers.NewRouter(handlers.New(store))
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
