package handlers

import (
	"time"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/storage"
)

// API takes dollars on the wire (client ergonomics) but stores cents (precision).
// The DTO layer is where that conversion lives.

type createEmployeeRequest struct {
	Name       string  `json:"name" binding:"required"`
	HourlyRate float64 `json:"hourly_rate" binding:"gt=0"`
}

type employeeResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	HourlyRate float64   `json:"hourly_rate"`
	CreatedAt  time.Time `json:"created_at"`
}

func newEmployeeResponse(e storage.Employee) employeeResponse {
	return employeeResponse{
		ID:         e.ID,
		Name:       e.Name,
		HourlyRate: moneyToDollars(e.HourlyRate),
		CreatedAt:  e.CreatedAt,
	}
}

type calculatePayslipRequest struct {
	EmployeeID  int64   `json:"employee_id" binding:"required"`
	HoursWorked float64 `json:"hours_worked" binding:"gte=0"`
}

type payslipResponse struct {
	ID             int64     `json:"id"`
	EmployeeID     int64     `json:"employee_id"`
	HoursWorked    float64   `json:"hours_worked"`
	GrossPay       float64   `json:"gross_pay"`
	FederalTax     float64   `json:"federal_tax"`
	SocialSecurity float64   `json:"social_security"`
	Medicare       float64   `json:"medicare"`
	NetPay         float64   `json:"net_pay"`
	CreatedAt      time.Time `json:"created_at"`
}

func newPayslipResponse(p storage.Payslip) payslipResponse {
	return payslipResponse{
		ID:             p.ID,
		EmployeeID:     p.EmployeeID,
		HoursWorked:    p.HoursWorked,
		GrossPay:       moneyToDollars(p.GrossPay),
		FederalTax:     moneyToDollars(p.FederalTax),
		SocialSecurity: moneyToDollars(p.SocialSecurity),
		Medicare:       moneyToDollars(p.Medicare),
		NetPay:         moneyToDollars(p.NetPay),
		CreatedAt:      p.CreatedAt,
	}
}

func moneyToDollars(m payroll.Money) float64 {
	return float64(m) / 100
}
