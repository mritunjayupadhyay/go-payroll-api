package storage

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
)

// Storage tests run against a real Postgres. They are skipped under
// `go test -short` and when no test DSN is reachable, so the default
// `go test ./...` stays fast and offline-friendly.
//
// To run them:
//
//	docker exec payroll-db psql -U postgres -c 'CREATE DATABASE payroll_test;'
//	TEST_DATABASE_URL='postgres://postgres:password@localhost:5432/payroll_test?sslmode=disable' \
//	  go test ./internal/storage

const defaultTestDSN = "postgres://postgres:password@localhost:5432/payroll_test?sslmode=disable"

var testStore *PostgresStorage

func TestMain(m *testing.M) {
	// Parse flags so testing.Short() reflects -short before m.Run().
	flag.Parse()

	if !testing.Short() {
		dsn := os.Getenv("TEST_DATABASE_URL")
		if dsn == "" {
			dsn = defaultTestDSN
		}
		ctx := context.Background()
		s, err := New(ctx, dsn)
		if err != nil {
			log.Printf("integration tests will skip: %v", err)
		} else if err := setupSchema(ctx, s); err != nil {
			log.Printf("integration tests will skip: setup schema: %v", err)
			s.Close()
		} else {
			testStore = s
			defer s.Close()
		}
	}

	os.Exit(m.Run())
}

func setupSchema(ctx context.Context, s *PostgresStorage) error {
	const drop = `DROP TABLE IF EXISTS payslips, employees CASCADE`
	if _, err := s.pool.Exec(ctx, drop); err != nil {
		return err
	}
	migration, err := os.ReadFile("../../migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, string(migration))
	return err
}

// requireDB skips the test if integration setup didn't succeed and resets
// table state so tests are independent. RESTART IDENTITY rewinds the
// BIGSERIAL sequences so each test sees deterministic IDs starting at 1.
func requireDB(t *testing.T) {
	t.Helper()
	if testStore == nil {
		t.Skip("integration tests disabled (no DB or -short)")
	}
	_, err := testStore.pool.Exec(context.Background(),
		"TRUNCATE payslips, employees RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func TestPostgres_CreateAndGetEmployee(t *testing.T) {
	requireDB(t)
	ctx := context.Background()

	created, err := testStore.CreateEmployee(ctx, "Jane", payroll.MoneyFromDollars(25))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}

	got, err := testStore.GetEmployee(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Jane" {
		t.Errorf("Name = %q, want %q", got.Name, "Jane")
	}
	if got.HourlyRate != 2500 {
		t.Errorf("HourlyRate = %d, want 2500", got.HourlyRate)
	}
}

func TestPostgres_GetEmployee_NotFound(t *testing.T) {
	requireDB(t)
	_, err := testStore.GetEmployee(context.Background(), 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestPostgres_SaveAndGetPayslip(t *testing.T) {
	requireDB(t)
	ctx := context.Background()

	emp, err := testStore.CreateEmployee(ctx, "Jane", payroll.MoneyFromDollars(25))
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}

	in := Payslip{
		EmployeeID:     emp.ID,
		HoursWorked:    40,
		GrossPay:       payroll.MoneyFromDollars(1000),
		FederalTax:     payroll.MoneyFromDollars(120),
		SocialSecurity: payroll.MoneyFromDollars(62),
		Medicare:       payroll.MoneyFromDollars(14.50),
		NetPay:         payroll.MoneyFromDollars(803.50),
	}
	saved, err := testStore.SavePayslip(ctx, in)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.ID == 0 || saved.CreatedAt.IsZero() {
		t.Fatalf("saved fields zero: %+v", saved)
	}

	got, err := testStore.GetPayslip(ctx, saved.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	// Compare the value-bearing fields. ID and CreatedAt are storage-assigned.
	if got.EmployeeID != in.EmployeeID ||
		got.HoursWorked != in.HoursWorked ||
		got.GrossPay != in.GrossPay ||
		got.FederalTax != in.FederalTax ||
		got.SocialSecurity != in.SocialSecurity ||
		got.Medicare != in.Medicare ||
		got.NetPay != in.NetPay {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, in)
	}
}

func TestPostgres_GetPayslip_NotFound(t *testing.T) {
	requireDB(t)
	_, err := testStore.GetPayslip(context.Background(), 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestPostgres_ListPayslipsForEmployee(t *testing.T) {
	requireDB(t)
	ctx := context.Background()

	emp, err := testStore.CreateEmployee(ctx, "Jane", payroll.MoneyFromDollars(25))
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}

	// Empty list before any payslips exist.
	got, err := testStore.ListPayslipsForEmployee(ctx, emp.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 payslips, got %d", len(got))
	}

	// Insert three; expect them back in created_at DESC order.
	for i := 0; i < 3; i++ {
		_, err := testStore.SavePayslip(ctx, Payslip{
			EmployeeID:  emp.ID,
			HoursWorked: float64(40 + i),
			GrossPay:    payroll.MoneyFromDollars(float64(1000 + i)),
		})
		if err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}

	got, err = testStore.ListPayslipsForEmployee(ctx, emp.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].CreatedAt.Before(got[i].CreatedAt) {
			t.Errorf("rows not in DESC order at index %d", i)
		}
	}
}

func TestPostgres_SavePayslip_UnknownEmployeeFails(t *testing.T) {
	requireDB(t)
	_, err := testStore.SavePayslip(context.Background(), Payslip{
		EmployeeID: 99999,
		GrossPay:   1,
	})
	if err == nil {
		t.Fatal("expected FK violation error, got nil")
	}
}
