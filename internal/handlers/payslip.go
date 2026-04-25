package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/payroll"
)

func CalculatePayslip(c *gin.Context) {
	var req calculatePayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slip := payroll.Calculate(req.toEmployee())
	c.JSON(http.StatusOK, newPayslipResponse(slip))
}
