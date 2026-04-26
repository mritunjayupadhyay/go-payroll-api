package handlers

import (
	"context"
	"io"
	"log/slog"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/storage"
)

// payslipStore is the consumer-side view of what handlers need from a store.
// Defined here, not in the storage package, so handlers depend on the
// behaviors they actually use — not on the full *PostgresStorage surface.
// Tests provide an in-memory fake that implements this interface.
type payslipStore interface {
	CreateEmployee(ctx context.Context, name string, hourlyRate payroll.Money) (storage.Employee, error)
	GetEmployee(ctx context.Context, id int64) (storage.Employee, error)
	SavePayslip(ctx context.Context, p storage.Payslip) (storage.Payslip, error)
	GetPayslip(ctx context.Context, id int64) (storage.Payslip, error)
	ListPayslipsForEmployee(ctx context.Context, employeeID int64) ([]storage.Payslip, error)
}

// API groups the HTTP handlers and their shared dependencies.
type API struct {
	store  payslipStore
	logger *slog.Logger
}

// New returns an API backed by the given store. Any value satisfying the
// payslipStore interface works — *storage.PostgresStorage in production,
// an in-memory fake in tests. A nil logger falls back to a discard logger
// so tests don't spam stdout.
func New(store payslipStore, logger *slog.Logger) *API {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	return &API{store: store, logger: logger}
}
