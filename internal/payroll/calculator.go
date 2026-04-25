// Package payroll computes gross-to-net payslips. It has no I/O dependencies
// and is intended to be unit-tested in isolation.
package payroll

import (
	"fmt"
	"math"
)

const (
	federalTaxRate     = 0.12
	socialSecurityRate = 0.062
	medicareRate       = 0.0145
)

// Money is an amount in cents. Using an integer type avoids the precision
// drift that float64 introduces under repeated multiplication and addition.
type Money int64

// MoneyFromDollars converts a dollar value to Money, rounding to the nearest cent.
func MoneyFromDollars(d float64) Money {
	return Money(math.Round(d * 100))
}

func (m Money) String() string {
	cents := int64(m)
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", sign, cents/100, cents%100)
}

type Employee struct {
	Name        string
	HourlyRate  Money
	HoursWorked float64
}

type Payslip struct {
	EmployeeName   string
	GrossPay       Money
	FederalTax     Money
	SocialSecurity Money
	Medicare       Money
	NetPay         Money
}

func Calculate(e Employee) Payslip {
	gross := Money(math.Round(float64(e.HourlyRate) * e.HoursWorked))
	federal := applyRate(gross, federalTaxRate)
	ss := applyRate(gross, socialSecurityRate)
	medicare := applyRate(gross, medicareRate)
	return Payslip{
		EmployeeName:   e.Name,
		GrossPay:       gross,
		FederalTax:     federal,
		SocialSecurity: ss,
		Medicare:       medicare,
		NetPay:         gross - federal - ss - medicare,
	}
}

func applyRate(m Money, rate float64) Money {
	return Money(math.Round(float64(m) * rate))
}

func (p Payslip) String() string {
	return fmt.Sprintf(
		"Payslip for %s\n  Gross:           %s\n  Federal Tax:     %s\n  Social Security: %s\n  Medicare:        %s\n  Net:             %s",
		p.EmployeeName, p.GrossPay, p.FederalTax, p.SocialSecurity, p.Medicare, p.NetPay,
	)
}
