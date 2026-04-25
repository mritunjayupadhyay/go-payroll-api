package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestCalculatePayslip_HappyPath(t *testing.T) {
	body := `{"name":"Jane","hourly_rate":25,"hours_worked":40}`

	r := NewRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/payslips/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var got payslipResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	want := payslipResponse{
		EmployeeName:   "Jane",
		GrossPay:       1000.00,
		FederalTax:     120.00,
		SocialSecurity: 62.00,
		Medicare:       14.50,
		NetPay:         803.50,
	}
	if got != want {
		t.Errorf("response = %+v\nwant      %+v", got, want)
	}
}

func TestCalculatePayslip_Validation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":"","hourly_rate":25,"hours_worked":40}`},
		{"missing name", `{"hourly_rate":25,"hours_worked":40}`},
		{"zero rate", `{"name":"Jane","hourly_rate":0,"hours_worked":40}`},
		{"negative rate", `{"name":"Jane","hourly_rate":-5,"hours_worked":40}`},
		{"negative hours", `{"name":"Jane","hourly_rate":25,"hours_worked":-1}`},
		{"malformed JSON", `{not json`},
		{"empty body", ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/payslips/calculate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
			}

			var resp map[string]string
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp["error"] == "" {
				t.Error("expected non-empty error field in response")
			}
		})
	}
}

func TestCalculatePayslip_ZeroHoursAllowed(t *testing.T) {
	body := `{"name":"Idle","hourly_rate":50,"hours_worked":0}`

	r := NewRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/payslips/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var got payslipResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.NetPay != 0 || got.GrossPay != 0 {
		t.Errorf("expected zero pay, got %+v", got)
	}
}
