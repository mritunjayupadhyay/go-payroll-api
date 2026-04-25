package handlers

import (
	"context"
	"sync"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/storage"
)

// fakeStore is an in-memory payslipStore for handler tests. CLAUDE.md
// guidance: don't mock what you can fake — a real implementation against a
// map is more honest than a generated mock.
type fakeStore struct {
	mu        sync.Mutex
	employees map[int64]storage.Employee
	payslips  map[int64]storage.Payslip
	nextEmpID int64
	nextPayID int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		employees: map[int64]storage.Employee{},
		payslips:  map[int64]storage.Payslip{},
	}
}

func (f *fakeStore) CreateEmployee(_ context.Context, name string, rate payroll.Money) (storage.Employee, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextEmpID++
	e := storage.Employee{ID: f.nextEmpID, Name: name, HourlyRate: rate}
	f.employees[e.ID] = e
	return e, nil
}

func (f *fakeStore) GetEmployee(_ context.Context, id int64) (storage.Employee, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.employees[id]
	if !ok {
		return storage.Employee{}, storage.ErrNotFound
	}
	return e, nil
}

func (f *fakeStore) SavePayslip(_ context.Context, p storage.Payslip) (storage.Payslip, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextPayID++
	p.ID = f.nextPayID
	f.payslips[p.ID] = p
	return p, nil
}

func (f *fakeStore) GetPayslip(_ context.Context, id int64) (storage.Payslip, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.payslips[id]
	if !ok {
		return storage.Payslip{}, storage.ErrNotFound
	}
	return p, nil
}

func (f *fakeStore) ListPayslipsForEmployee(_ context.Context, employeeID int64) ([]storage.Payslip, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []storage.Payslip
	for _, p := range f.payslips {
		if p.EmployeeID == employeeID {
			out = append(out, p)
		}
	}
	return out, nil
}
