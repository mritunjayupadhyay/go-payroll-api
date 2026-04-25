// Package handlers wires HTTP routes onto a Gin engine and exposes the
// handler functions that translate between JSON DTOs and the payroll
// domain types.
package handlers

import "github.com/gin-gonic/gin"

// NewRouter returns a fully configured Gin engine. Tests use this with
// httptest; cmd/server uses it as the application root.
func NewRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", Health)

	api := r.Group("/api")
	{
		api.POST("/payslips/calculate", CalculatePayslip)
	}

	return r
}
