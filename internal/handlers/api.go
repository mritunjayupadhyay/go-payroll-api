package handlers

import (
	"context"

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
	store payslipStore
}

// New returns an API backed by the given store. Any value satisfying the
// payslipStore interface works — *storage.PostgresStorage in production,
// an in-memory fake in tests.
func New(store payslipStore) *API {
	return &API{store: store}
}
