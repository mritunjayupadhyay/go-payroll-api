package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
)

func main() {
	name := flag.String("name", "", "employee name")
	rate := flag.Float64("rate", 0, "hourly rate in dollars")
	hours := flag.Float64("hours", 0, "hours worked")
	flag.Parse()

	if err := validate(*name, *rate, *hours); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	slip := payroll.Calculate(payroll.Employee{
		Name:        *name,
		HourlyRate:  payroll.MoneyFromDollars(*rate),
		HoursWorked: *hours,
	})
	fmt.Println(slip)
}

func validate(name string, rate, hours float64) error {
	if name == "" {
		return errors.New("name is required")
	}
	if rate <= 0 {
		return errors.New("rate must be greater than 0")
	}
	if hours < 0 {
		return errors.New("hours must not be negative")
	}
	return nil
}
