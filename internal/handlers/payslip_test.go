package handlers

import (
	"encoding/json"
	"fmt"
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

// newTestRouter wires a fresh fake store + router so each test is isolated.
func newTestRouter() (*gin.Engine, *fakeStore) {
	store := newFakeStore()
	return NewRouter(New(store, nil)), store
}

func doJSON(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestCreateEmployee_HappyPath(t *testing.T) {
	r, _ := newTestRouter()
	w := doJSON(t, r, http.MethodPost, "/api/employees", `{"name":"Jane","hourly_rate":25}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", w.Code, w.Body.String())
	}
	var got employeeResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID == 0 || got.Name != "Jane" || got.HourlyRate != 25 {
		t.Errorf("response = %+v", got)
	}
}

func TestCreateEmployee_Validation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing name", `{"hourly_rate":25}`},
		{"empty name", `{"name":"","hourly_rate":25}`},
		{"zero rate", `{"name":"Jane","hourly_rate":0}`},
		{"negative rate", `{"name":"Jane","hourly_rate":-1}`},
		{"malformed JSON", `{not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := newTestRouter()
			w := doJSON(t, r, http.MethodPost, "/api/employees", tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body = %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestCalculatePayslip_HappyPath(t *testing.T) {
	r, _ := newTestRouter()

	// Seed an employee.
	w := doJSON(t, r, http.MethodPost, "/api/employees", `{"name":"Jane","hourly_rate":25}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("seed employee: status %d", w.Code)
	}
	var emp employeeResponse
	_ = json.NewDecoder(w.Body).Decode(&emp)

	body := fmt.Sprintf(`{"employee_id":%d,"hours_worked":40}`, emp.ID)
	w = doJSON(t, r, http.MethodPost, "/api/payslips/calculate", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", w.Code, w.Body.String())
	}
	var got payslipResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := payslipResponse{
		ID:             got.ID,         // generated, accept whatever
		EmployeeID:     emp.ID,
		HoursWorked:    40,
		GrossPay:       1000.00,
		FederalTax:     120.00,
		SocialSecurity: 62.00,
		Medicare:       14.50,
		NetPay:         803.50,
		CreatedAt:      got.CreatedAt, // fake doesn't set, but accept zero
	}
	if got != want {
		t.Errorf("response = %+v\nwant      %+v", got, want)
	}
}

func TestCalculatePayslip_UnknownEmployee(t *testing.T) {
	r, _ := newTestRouter()
	w := doJSON(t, r, http.MethodPost, "/api/payslips/calculate", `{"employee_id":999,"hours_worked":40}`)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404; body = %s", w.Code, w.Body.String())
	}
}

func TestCalculatePayslip_Validation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing employee_id", `{"hours_worked":40}`},
		{"zero employee_id", `{"employee_id":0,"hours_worked":40}`},
		{"negative hours", `{"employee_id":1,"hours_worked":-1}`},
		{"malformed JSON", `{not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := newTestRouter()
			w := doJSON(t, r, http.MethodPost, "/api/payslips/calculate", tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body = %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestGetPayslip_NotFound(t *testing.T) {
	r, _ := newTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/payslips/999", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestListPayslipsForEmployee_Empty(t *testing.T) {
	r, _ := newTestRouter()

	w := doJSON(t, r, http.MethodPost, "/api/employees", `{"name":"Solo","hourly_rate":50}`)
	var emp employeeResponse
	_ = json.NewDecoder(w.Body).Decode(&emp)

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/employees/%d/payslips", emp.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	// Must be `[]` not `null`.
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Errorf("body = %q, want []", w.Body.String())
	}
}

func TestListPayslipsForEmployee_ReturnsRows(t *testing.T) {
	r, _ := newTestRouter()

	w := doJSON(t, r, http.MethodPost, "/api/employees", `{"name":"Jane","hourly_rate":25}`)
	var emp employeeResponse
	_ = json.NewDecoder(w.Body).Decode(&emp)

	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"employee_id":%d,"hours_worked":40}`, emp.ID)
		w := doJSON(t, r, http.MethodPost, "/api/payslips/calculate", body)
		if w.Code != http.StatusCreated {
			t.Fatalf("seed payslip %d: status %d", i, w.Code)
		}
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/employees/%d/payslips", emp.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got []payslipResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
}
