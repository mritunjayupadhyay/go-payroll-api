// Package storage provides PostgreSQL-backed persistence for employees and
// payslips. The package exposes a concrete *PostgresStorage; consumers
// (handlers) should declare their own minimal interface against it.
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
)

// ErrNotFound is returned when a row is not found. Callers can check it with
// errors.Is(err, storage.ErrNotFound) and map it to a 404.
var ErrNotFound = errors.New("not found")

type Employee struct {
	ID         int64
	Name       string
	HourlyRate payroll.Money
	CreatedAt  time.Time
}

type Payslip struct {
	ID             int64
	EmployeeID     int64
	HoursWorked    float64
	GrossPay       payroll.Money
	FederalTax     payroll.Money
	SocialSecurity payroll.Money
	Medicare       payroll.Money
	NetPay         payroll.Money
	CreatedAt      time.Time
}

type PostgresStorage struct {
	pool *pgxpool.Pool
}

// New connects to Postgres at dsn and verifies the connection with a ping.
// The caller owns the returned *PostgresStorage and must call Close.
func New(ctx context.Context, dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}

func (s *PostgresStorage) CreateEmployee(ctx context.Context, name string, hourlyRate payroll.Money) (Employee, error) {
	const q = `INSERT INTO employees (name, hourly_rate)
	           VALUES ($1, $2)
	           RETURNING id, created_at`
	e := Employee{Name: name, HourlyRate: hourlyRate}
	if err := s.pool.QueryRow(ctx, q, name, int64(hourlyRate)).Scan(&e.ID, &e.CreatedAt); err != nil {
		return Employee{}, fmt.Errorf("inserting employee: %w", err)
	}
	return e, nil
}

func (s *PostgresStorage) GetEmployee(ctx context.Context, id int64) (Employee, error) {
	const q = `SELECT id, name, hourly_rate, created_at
	           FROM employees
	           WHERE id = $1`
	var (
		e    Employee
		rate int64
	)
	err := s.pool.QueryRow(ctx, q, id).Scan(&e.ID, &e.Name, &rate, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Employee{}, ErrNotFound
	}
	if err != nil {
		return Employee{}, fmt.Errorf("getting employee %d: %w", id, err)
	}
	e.HourlyRate = payroll.Money(rate)
	return e, nil
}

func (s *PostgresStorage) SavePayslip(ctx context.Context, p Payslip) (Payslip, error) {
	const q = `INSERT INTO payslips
	             (employee_id, hours_worked, gross_pay, federal_tax, social_security, medicare, net_pay)
	           VALUES ($1, $2, $3, $4, $5, $6, $7)
	           RETURNING id, created_at`
	err := s.pool.QueryRow(ctx, q,
		p.EmployeeID,
		p.HoursWorked,
		int64(p.GrossPay),
		int64(p.FederalTax),
		int64(p.SocialSecurity),
		int64(p.Medicare),
		int64(p.NetPay),
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return Payslip{}, fmt.Errorf("inserting payslip: %w", err)
	}
	return p, nil
}

func (s *PostgresStorage) GetPayslip(ctx context.Context, id int64) (Payslip, error) {
	const q = `SELECT id, employee_id, hours_worked,
	                  gross_pay, federal_tax, social_security, medicare, net_pay,
	                  created_at
	           FROM payslips
	           WHERE id = $1`
	p, err := scanPayslip(s.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Payslip{}, ErrNotFound
	}
	if err != nil {
		return Payslip{}, fmt.Errorf("getting payslip %d: %w", id, err)
	}
	return p, nil
}

func (s *PostgresStorage) ListPayslipsForEmployee(ctx context.Context, employeeID int64) ([]Payslip, error) {
	const q = `SELECT id, employee_id, hours_worked,
	                  gross_pay, federal_tax, social_security, medicare, net_pay,
	                  created_at
	           FROM payslips
	           WHERE employee_id = $1
	           ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q, employeeID)
	if err != nil {
		return nil, fmt.Errorf("querying payslips for employee %d: %w", employeeID, err)
	}
	defer rows.Close()

	var result []Payslip
	for rows.Next() {
		p, err := scanPayslip(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning payslip row: %w", err)
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating payslip rows: %w", err)
	}
	return result, nil
}

// scanRow is the minimal subset of pgx.Row and pgx.Rows we need so the same
// scan logic works for both single-row and multi-row queries.
type scanRow interface {
	Scan(dest ...any) error
}

func scanPayslip(r scanRow) (Payslip, error) {
	var (
		p                          Payslip
		gross, fed, ss, med, net int64
	)
	if err := r.Scan(
		&p.ID, &p.EmployeeID, &p.HoursWorked,
		&gross, &fed, &ss, &med, &net,
		&p.CreatedAt,
	); err != nil {
		return Payslip{}, err
	}
	p.GrossPay = payroll.Money(gross)
	p.FederalTax = payroll.Money(fed)
	p.SocialSecurity = payroll.Money(ss)
	p.Medicare = payroll.Money(med)
	p.NetPay = payroll.Money(net)
	return p, nil
}
