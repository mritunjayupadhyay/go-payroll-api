package handlers

import "github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"

// calculatePayslipRequest is the wire shape for POST /api/payslips/calculate.
// It is intentionally separate from payroll.Employee: the domain type uses
// Money (cents) for precision; the API takes dollars for client ergonomics.
type calculatePayslipRequest struct {
	Name        string  `json:"name" binding:"required"`
	HourlyRate  float64 `json:"hourly_rate" binding:"gt=0"`
	HoursWorked float64 `json:"hours_worked" binding:"gte=0"`
}

type payslipResponse struct {
	EmployeeName   string  `json:"employee_name"`
	GrossPay       float64 `json:"gross_pay"`
	FederalTax     float64 `json:"federal_tax"`
	SocialSecurity float64 `json:"social_security"`
	Medicare       float64 `json:"medicare"`
	NetPay         float64 `json:"net_pay"`
}

func (r calculatePayslipRequest) toEmployee() payroll.Employee {
	return payroll.Employee{
		Name:        r.Name,
		HourlyRate:  payroll.MoneyFromDollars(r.HourlyRate),
		HoursWorked: r.HoursWorked,
	}
}

func newPayslipResponse(p payroll.Payslip) payslipResponse {
	return payslipResponse{
		EmployeeName:   p.EmployeeName,
		GrossPay:       moneyToDollars(p.GrossPay),
		FederalTax:     moneyToDollars(p.FederalTax),
		SocialSecurity: moneyToDollars(p.SocialSecurity),
		Medicare:       moneyToDollars(p.Medicare),
		NetPay:         moneyToDollars(p.NetPay),
	}
}

func moneyToDollars(m payroll.Money) float64 {
	return float64(m) / 100
}
