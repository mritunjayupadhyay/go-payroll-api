package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
	"github.com/mritunjayupadhyay/go-payroll-api/internal/storage"
)

func (a *API) CalculatePayslip(c *gin.Context) {
	var req calculatePayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	emp, err := a.store.GetEmployee(ctx, req.EmployeeID)
	if errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	calc := payroll.Calculate(payroll.Employee{
		Name:        emp.Name,
		HourlyRate:  emp.HourlyRate,
		HoursWorked: req.HoursWorked,
	})

	saved, err := a.store.SavePayslip(ctx, storage.Payslip{
		EmployeeID:     emp.ID,
		HoursWorked:    req.HoursWorked,
		GrossPay:       calc.GrossPay,
		FederalTax:     calc.FederalTax,
		SocialSecurity: calc.SocialSecurity,
		Medicare:       calc.Medicare,
		NetPay:         calc.NetPay,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newPayslipResponse(saved))
}

func (a *API) GetPayslip(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := a.store.GetPayslip(c.Request.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "payslip not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newPayslipResponse(p))
}

func (a *API) ListPayslipsForEmployee(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slips, err := a.store.ListPayslipsForEmployee(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Pre-allocate so a zero-row result encodes as `[]`, not `null`.
	resp := make([]payslipResponse, 0, len(slips))
	for _, p := range slips {
		resp = append(resp, newPayslipResponse(p))
	}
	c.JSON(http.StatusOK, resp)
}
