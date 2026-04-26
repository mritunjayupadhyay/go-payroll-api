// Package handlers wires HTTP routes onto a Gin engine and exposes the
// handler functions that translate between JSON DTOs and the payroll
// and storage domain types.
package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/mritunjayupadhyay/go-payroll-api/internal/middleware"
)

// NewRouter returns a fully configured Gin engine. Tests use this with
// httptest; cmd/server uses it as the application root.
//
// gin.New (not gin.Default) so we pick the middleware: structured access
// logs via slog, plus Gin's recovery to convert panics into 500s.
func NewRouter(api *API) *gin.Engine {
	r := gin.New()
	r.Use(middleware.AccessLog(api.logger), gin.Recovery())

	r.GET("/health", Health)

	g := r.Group("/api")
	{
		g.POST("/employees", api.CreateEmployee)
		g.GET("/employees/:id", api.GetEmployee)
		g.GET("/employees/:id/payslips", api.ListPayslipsForEmployee)

		g.POST("/payslips/calculate", api.CalculatePayslip)
		g.GET("/payslips/:id", api.GetPayslip)
	}

	return r
}
