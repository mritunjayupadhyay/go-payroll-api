# Build Plan — Payroll Service

This document tracks how the project is being built, milestone by milestone. Each milestone produces a working, demoable state so the project is presentable at any point — not just when "finished."

---

## Background

This is a personal learning project for a senior backend engineer (8+ years Node.js / NestJS / TypeScript) transitioning into Go. Goals, in priority order:

1. Learn idiomatic Go — write Go that a Go engineer would write, not Node.js patterns translated into Go syntax.
2. Build a portfolio-quality, demoable project for CV and interviews within ~7 days.
3. Follow industry-standard patterns (layered architecture, dependency injection, testing, containerization).

Time budget: ~2-3 hours/day, more on weekends. Total: ~15-20 hours over 7 days.

---

## Tech stack

- Go 1.22+
- Gin (`github.com/gin-gonic/gin`)
- PostgreSQL 16 with `github.com/jackc/pgx/v5/pgxpool`
- Docker, Docker Compose
- Standard library: `log/slog`, `testing`, `net/http/httptest`, `flag`

---

## Final project structure

```
/payroll-service
├── cmd/
│   ├── cli/main.go         # CLI tool
│   └── server/main.go      # HTTP server
├── internal/
│   ├── payroll/            # Calculation logic + tests
│   ├── handlers/           # HTTP handlers
│   └── storage/            # PostgreSQL access
├── migrations/
│   └── 001_init.sql
├── deploy/
│   └── k8s/                # Optional Kubernetes manifests
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
├── README.md
├── CLAUDE.md
└── BUILD_PLAN.md
```

---

## Milestone tracking

Mark milestones complete as you finish them.

- [ ] **M0** — Setup
- [ ] **M1** — CLI with core payroll calculation
- [ ] **M2** — REST API
- [ ] **M3** — PostgreSQL persistence
- [ ] **M4** — Docker
- [ ] **M5** — Polish (README, structured logging, CI)
- [ ] **M6** — Kubernetes (optional)

---

## Milestone 0 — Setup (1-2 hours)

**Goal:** Working Go environment, GitHub repo, hello world running.

### Steps

1. Install Go 1.22+ (`brew install go` on Mac, or download from go.dev).
2. Verify with `go version`.
3. Create a public GitHub repo named `payroll-service`. Initialize with README and Go `.gitignore` template (or use the prepared `.gitignore`).
4. Clone locally and `cd` into it.
5. Run `go mod init github.com/<my-github-username>/payroll-service`.
6. Create `main.go` with a basic hello world:
   ```go
   package main

   import "fmt"

   func main() {
       fmt.Println("hello world")
   }
   ```
7. Run with `go run main.go`. Confirm it prints.
8. Add the prepared `README.md`, `CLAUDE.md`, `BUILD_PLAN.md`, `.gitignore` files to the repo root.
9. Commit and push.
n
### Shippable state

Repo exists publicly, Go module initialized, code compiles.

---

## Milestone 1 — CLI with core payroll calculation (2-3 hours)

**Goal:** A command-line tool that computes gross-to-net pay. Pure logic, no HTTP, no database.

### Steps

1. Create folder structure:
   - `cmd/cli/main.go`
   - `internal/payroll/calculator.go`
   - `internal/payroll/calculator_test.go`
2. Delete the root `main.go` from M0 (the CLI now lives in `cmd/cli/`).
3. In `calculator.go`, define structs:
   - `Employee` with fields `Name`, `HourlyRate`, `HoursWorked`
   - `Payslip` with fields `EmployeeName`, `GrossPay`, `FederalTax`, `SocialSecurity`, `Medicare`, `NetPay`
4. Write a `Calculate(employee Employee) Payslip` function using mock tax rates:
   - Federal: 12% of gross
   - Social Security: 6.2% of gross
   - Medicare: 1.45% of gross
   - Net = gross minus all taxes
5. In `cmd/cli/main.go`, accept command-line flags using the `flag` package: name, hourly rate, hours. Call `Calculate` and print the payslip.
6. Run: `go run ./cmd/cli -name="John" -rate=25 -hours=40`. Should print a formatted payslip.
7. Write table-driven unit tests in `calculator_test.go`. Cover at least 3 cases:
   - Normal hours
   - Zero hours
   - Edge case with high rate
8. Run `go test ./...` and confirm passing.
9. Update README with how to build and run the CLI, with example output.
10. Commit and push.

### Shippable state

Working Go program with real payroll logic, unit tests, clean folder structure, documentation.

---

## Milestone 2 — REST API (2-3 hours)

**Goal:** Expose the calculation as an HTTP endpoint. The calculation logic from M1 stays unchanged.

### Steps

1. Add Gin: `go get github.com/gin-gonic/gin`.
2. Create `cmd/server/main.go` (the CLI from M1 keeps working alongside).
3. Build endpoints:
   - `GET /health` returning `{"status": "ok"}`
   - `POST /api/payslips/calculate` accepting JSON, returning JSON payslip
4. Keep handlers thin: validate input → call `payroll.Calculate` → return response. The `payroll` package does not change.
5. Add input validation (negative hours, zero rate, empty name) returning HTTP 400 with a useful error message.
6. Test manually with curl:
   ```bash
   curl -X POST http://localhost:8080/api/payslips/calculate \
     -H "Content-Type: application/json" \
     -d '{"name":"Jane","hourly_rate":25,"hours_worked":40}'
   ```
7. Write integration tests using `net/http/httptest`.
8. Update README with API documentation and curl examples.
9. Commit and push.

### Shippable state

REST API with real business logic, validation, tests, and documentation.

---

## Milestone 3 — PostgreSQL persistence (3-4 hours)

**Goal:** Store and retrieve payslips.

### Steps

1. Run Postgres locally with Docker:
   ```bash
   docker run --name payroll-db \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=payroll \
     -p 5432:5432 -d postgres:16
   ```
2. Add the pgx driver: `go get github.com/jackc/pgx/v5/pgxpool`.
3. Create `migrations/001_init.sql` with `employees` and `payslips` tables.
4. Create `internal/storage/postgres.go`:
   - Define a `Storage` interface so handlers depend on the interface, not the concrete type.
   - Implement methods: `SavePayslip`, `GetPayslip`, `ListPayslipsForEmployee`.
5. Add new endpoints:
   - `POST /api/payslips/calculate` now also saves the payslip and returns its ID
   - `GET /api/payslips/:id`
   - `GET /api/employees/:id/payslips`
6. Use environment variables for the database URL. Add `.env.example` with placeholder values. The real `.env` is already in `.gitignore`.
7. Add tests for the storage layer.
8. Update README with database setup instructions.
9. Commit and push.

### Shippable state

Full-stack Go service with persistence, multiple endpoints, and SQL migrations.

---

## Milestone 4 — Docker (1-2 hours)

**Goal:** One-command deployment.

### Steps

1. Create a multi-stage Dockerfile:
   - Build stage: `golang:1.22` image, copies source, runs `go build`
   - Runtime stage: alpine or distroless, copies just the binary
2. Create `docker-compose.yml` running both the app and Postgres together.
3. Test: `docker-compose up --build`. Confirm the API works through Docker.
4. Add a "Quick Start with Docker" section to README.
5. Commit and push.

### Shippable state

Anyone can clone and run the project with one command.

---

## Milestone 5 — Polish (2-3 hours)

**Goal:** Portfolio-quality repo.

### Steps

1. Rewrite README to include:
   - One-paragraph project description and why it was built
   - Tech stack
   - Architecture diagram (ASCII or Mermaid)
   - Quick start with Docker
   - API documentation with curl examples
   - Project structure
   - "What I Learned" section (idiomatic Go vs Node patterns, standard library, testing patterns)
2. Add structured logging using `log/slog`. Replace any `fmt.Println` debug statements.
3. Add request ID middleware so logs can be traced across a request.
4. Add `.github/workflows/ci.yml` running `go test ./...` on every push. Adds a CI badge to README.
5. Tag a release: `v0.1.0` with release notes.
6. Commit and push.

### Shippable state

Portfolio-quality project ready to link from CV.

---

## Milestone 6 — Kubernetes (optional, 2-3 hours)

**Goal:** End-to-end deployment with Kubernetes. Skip if running short on time — M1 through M5 is already strong.

### Steps

1. Install kind: `brew install kind`.
2. Create a cluster: `kind create cluster --name payroll`.
3. Create `deploy/k8s/` with manifests:
   - `deployment.yaml`
   - `service.yaml`
   - `configmap.yaml`
   - `secret.yaml` (template only — real secrets stay in environment)
   - `postgres.yaml` (Postgres in cluster, or document using external)
4. Build the Docker image and load it: `kind load docker-image payroll-service`.
5. Apply manifests: `kubectl apply -f deploy/k8s/`.
6. Verify: `kubectl get pods`, `kubectl logs`, `kubectl port-forward`.
7. Take screenshots and add a "Kubernetes deployment" section to README.

### Shippable state

End-to-end Go service with Kubernetes deployment. Significant credibility boost in interviews.

---

## How to use this plan with Claude Code

When starting a session, give Claude Code a clear, scoped instruction:

> "I'm working on Milestone 2. Read CLAUDE.md and BUILD_PLAN.md, look at what's currently in `internal/payroll/`, then propose the structure for the HTTP handlers before writing any code."

Vague openers ("help me with my Go project") get vague responses. Specific openers get focused, useful output.

---

## Pacing advice

A polished M4 beats a rushed M6. Don't rush through milestones if it means doing each one poorly.

If running short on time, here's where it's safe to stop:

- Stop at **M1**: "I'm exploring Go through a small CLI project." Honest and fine.
- Stop at **M2**: "I built a Go REST API for payroll calculations." Solid.
- Stop at **M3**: "I built a Go REST API with PostgreSQL persistence." Strong.
- Stop at **M4 or M5**: Genuinely impressive for a self-taught project.
- Reach **M6**: Can defend Go and Kubernetes credibly in interviews.

---

## What this project is NOT

- Not a real payroll system. Tax calculations are mock and intentionally simplified.
- Not a microservices showcase. Single service, deliberately.
- Not trying to demonstrate every Go feature. Goroutines, channels, generics — only used where genuinely the right tool.
- Not production-grade. No real auth, audit log, or compliance handling. Adding these would be scope creep.

---

## Things to avoid

- ORMs (GORM, etc.) — learning `database/sql` and `pgx` idioms directly.
- Java-style Go (excessive abstraction, factory patterns, deep inheritance simulation).
- Premature use of goroutines and channels — most code in this project is correctly synchronous.
- Premature optimization — clarity first, performance only if measurements justify it.
- Adding features from later milestones silently.