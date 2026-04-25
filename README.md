# Payroll Calculation Service

A REST API in Go that computes gross-to-net pay with mock tax withholding. Built as a hands-on project to deepen Go fluency and explore patterns common in financial calculation services.

> **Note:** This is a personal learning project, not a production payroll system. Tax calculations use simplified mock rates; real-world payroll requires jurisdiction-specific tax tables, compliance handling, and audit trails.

---

## Tech Stack

- **Language:** Go 1.22+
- **Web framework:** Gin
- **Database:** PostgreSQL 16
- **Containerization:** Docker, Docker Compose
- **Testing:** Go standard `testing` package, `httptest`
- **Logging:** `log/slog` (structured logging)

---

## What It Does

Given an employee's hourly rate and hours worked, the service:

1. Computes gross pay
2. Applies mock tax withholdings (federal, social security, medicare)
3. Returns a payslip with the breakdown
4. Persists the payslip to PostgreSQL
5. Allows retrieval of past payslips by ID or by employee

---

## Architecture

```
┌─────────────┐      ┌──────────────┐      ┌──────────────┐
│   Client    │─────▶│  Gin Router  │─────▶│   Handlers   │
└─────────────┘      └──────────────┘      └──────┬───────┘
                                                  │
                            ┌─────────────────────┼─────────────────────┐
                            ▼                     ▼                     ▼
                    ┌───────────────┐     ┌──────────────┐     ┌───────────────┐
                    │   Calculator  │     │   Storage    │     │   Validation  │
                    │   (payroll)   │     │  (postgres)  │     │   (input)     │
                    └───────────────┘     └──────┬───────┘     └───────────────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │  PostgreSQL  │
                                          └──────────────┘
```

The project follows a layered architecture: HTTP handlers are thin and delegate to the `payroll` package (business logic) and the `storage` package (persistence). The calculator has no knowledge of HTTP or the database, making it easy to test in isolation.

---

## Project Structure

```
.
├── cmd/
│   ├── cli/              # Command-line interface (Milestone 1)
│   │   └── main.go
│   └── server/           # HTTP server (Milestone 2+)
│       └── main.go
├── internal/
│   ├── payroll/          # Core calculation logic
│   │   ├── calculator.go
│   │   └── calculator_test.go
│   ├── handlers/         # HTTP handlers
│   │   └── payslips.go
│   └── storage/          # Database access
│       └── postgres.go
├── migrations/
│   └── 001_init.sql      # Database schema
├── deploy/
│   ├── docker-compose.yml
│   └── k8s/              # Kubernetes manifests (optional)
├── Dockerfile
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

---

## Quick Start (Docker)

The fastest way to run the service:

```bash
git clone https://github.com/yourusername/payroll-service.git
cd payroll-service
docker-compose up --build
```

The API will be available at `http://localhost:8080`.

---

## Local Development

### Prerequisites

- Go 1.22 or later
- PostgreSQL 16 (or run via Docker)
- Make (optional, for convenience commands)

### Setup

```bash
# Install dependencies
go mod download

# Copy environment template
cp .env.example .env
# Edit .env with your database credentials

# Run database (via Docker)
docker run --name payroll-db \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=payroll \
  -p 5432:5432 \
  -d postgres:16

# Run migrations
psql -h localhost -U postgres -d payroll -f migrations/001_init.sql

# Run the server
go run ./cmd/server
```

### Running tests

```bash
go test ./...
go test -v ./internal/payroll  # Verbose, single package
go test -cover ./...           # With coverage
```

---

## API Reference

### Health check

```bash
curl http://localhost:8080/health
```

Response:
```json
{ "status": "ok" }
```

### Calculate a payslip

```bash
curl -X POST http://localhost:8080/api/payslips/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "hourly_rate": 25.00,
    "hours_worked": 40
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "employee_name": "Jane Doe",
  "gross_pay": 1000.00,
  "federal_tax": 120.00,
  "social_security": 62.00,
  "medicare": 14.50,
  "net_pay": 803.50,
  "calculated_at": "2026-04-25T10:30:00Z"
}
```

### Retrieve a payslip

```bash
curl http://localhost:8080/api/payslips/{id}
```

### List payslips for an employee

```bash
curl http://localhost:8080/api/employees/{id}/payslips
```

---

## Tax Calculation (Mock)

The service uses simplified, illustrative tax rates. **These are not real tax rates** and should not be used for actual payroll.

| Tax Type        | Rate    |
|-----------------|---------|
| Federal         | 12.00%  |
| Social Security | 6.20%   |
| Medicare        | 1.45%   |

A real payroll system would account for filing status, exemptions, year-to-date earnings, jurisdiction-specific rules, pre-tax deductions, and many other factors.

---

## CLI Usage

The project ships a small CLI that runs the calculation locally — no HTTP, no database. Useful as a quick sanity check and as a demonstration of the core logic in isolation.

### Run directly with `go run`

```bash
go run ./cmd/cli -name="Jane" -rate=25 -hours=40
```

Output:
```
Payslip for Jane
  Gross:           $1000.00
  Federal Tax:     $120.00
  Social Security: $62.00
  Medicare:        $14.50
  Net:             $803.50
```

### Build a standalone binary

Go compiles to a single static binary with no runtime dependency.

```bash
go build -o bin/payroll-cli ./cmd/cli
./bin/payroll-cli -name="Jane" -rate=25 -hours=40
```

The binary is self-contained — copy it anywhere and run it.

### Flags

| Flag      | Type    | Description                       |
|-----------|---------|-----------------------------------|
| `-name`   | string  | Employee name (required)          |
| `-rate`   | float64 | Hourly rate in dollars (required, > 0) |
| `-hours`  | float64 | Hours worked (required, ≥ 0)      |

### Validation errors

Bad input is rejected with a stderr message and exit code 1:

```bash
$ go run ./cmd/cli -rate=25 -hours=40
error: name is required
$ echo $?
1
```

### Running the tests

```bash
go test ./...                   # all packages
go test -v ./internal/payroll   # verbose, calculator only
go test -cover ./...            # with coverage
```

---

## What I Learned

> _Fill this section in as you build the project — interviewers love it._

- **Idiomatic Go vs. Node patterns:** [e.g., explicit error returns instead of try/catch, struct composition instead of class inheritance, interfaces being implicitly satisfied]
- **Concurrency:** [if you use goroutines anywhere, note what you learned]
- **Standard library depth:** [Go's stdlib covers a lot of ground that requires libraries in Node — note what you found useful]
- **Build and deployment:** [single binary output, multi-stage Docker builds, image size]
- **Testing:** [table-driven tests, the testing package, httptest]

---

## Roadmap

- [x] Core calculation logic with unit tests
- [x] REST API with input validation
- [x] PostgreSQL persistence
- [x] Dockerized deployment
- [ ] Kubernetes manifests for local cluster
- [ ] Structured logging with request tracing
- [ ] Configurable tax rates via environment / config file
- [ ] Authentication (JWT)
- [ ] Rate limiting

---

## License

MIT — feel free to use this as a reference for your own learning.

---

## About

Built by [Mritunjay Upadhyay](https://github.com/mritunjayupadhyay) — Senior Full-Stack Engineer based in Bangkok. Reach me at mupadhyay00@gmail.com.