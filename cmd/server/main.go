package main

import (
	"log"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/handlers"
)

func main() {
	r := handlers.NewRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
