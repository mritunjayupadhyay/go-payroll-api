package payroll

import "testing"

func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		employee Employee
		want     Payslip
	}{
		{
			name:     "normal hours",
			employee: Employee{Name: "Jane", HourlyRate: MoneyFromDollars(25), HoursWorked: 40},
			want: Payslip{
				EmployeeName:   "Jane",
				GrossPay:       100000,
				FederalTax:     12000,
				SocialSecurity: 6200,
				Medicare:       1450,
				NetPay:         80350,
			},
		},
		{
			name:     "zero hours",
			employee: Employee{Name: "Idle", HourlyRate: MoneyFromDollars(50), HoursWorked: 0},
			want:     Payslip{EmployeeName: "Idle"},
		},
		{
			name:     "high rate",
			employee: Employee{Name: "CEO", HourlyRate: MoneyFromDollars(500), HoursWorked: 40},
			want: Payslip{
				EmployeeName:   "CEO",
				GrossPay:       2000000,
				FederalTax:     240000,
				SocialSecurity: 124000,
				Medicare:       29000,
				NetPay:         1607000,
			},
		},
		{
			name:     "fractional hours",
			employee: Employee{Name: "PartTime", HourlyRate: MoneyFromDollars(20), HoursWorked: 37.5},
			want: Payslip{
				EmployeeName:   "PartTime",
				GrossPay:       75000,
				FederalTax:     9000,
				SocialSecurity: 4650,
				Medicare:       1088,
				NetPay:         60262,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.employee)
			if got != tt.want {
				t.Errorf("Calculate() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		m    Money
		want string
	}{
		{0, "$0.00"},
		{100, "$1.00"},
		{12345, "$123.45"},
		{5, "$0.05"},
		{-500, "-$5.00"},
	}
	for _, tt := range tests {
		if got := tt.m.String(); got != tt.want {
			t.Errorf("Money(%d).String() = %q, want %q", tt.m, got, tt.want)
		}
	}
}

func TestMoneyFromDollarsRounds(t *testing.T) {
	// Note: float64 cannot represent every decimal exactly (e.g. 1.005 is
	// actually 1.00499...), so MoneyFromDollars only promises rounding
	// behaviour for dollar amounts that float64 can represent cleanly.
	tests := []struct {
		dollars float64
		want    Money
	}{
		{0.0, 0},
		{1.00, 100},
		{1.004, 100},
		{1.006, 101},
		{12.50, 1250},
	}
	for _, tt := range tests {
		if got := MoneyFromDollars(tt.dollars); got != tt.want {
			t.Errorf("MoneyFromDollars(%v) = %d, want %d", tt.dollars, got, tt.want)
		}
	}
}
