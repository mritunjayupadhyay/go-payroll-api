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
│   ├── handlers/         # HTTP handlers (Gin)
│   │   ├── api.go            # API struct + consumer-side store interface
│   │   ├── router.go
│   │   ├── employee.go
│   │   ├── payslip.go
│   │   ├── health.go
│   │   ├── dto.go            # Request/response DTOs + dollar↔cent conversion
│   │   ├── *_test.go
│   │   └── store_fake_test.go # In-memory fake for handler tests
│   └── storage/          # PostgreSQL persistence (pgx/v5)
│       ├── postgres.go
│       └── postgres_test.go  # Integration tests, gated by -short / TEST_DATABASE_URL
├── migrations/
│   └── 001_init.sql      # Database schema (also auto-applied on first DB boot via compose)
├── deploy/
│   └── k8s/              # Kubernetes manifests (optional, Milestone 6)
├── Dockerfile            # Multi-stage build: golang:1.26-alpine → distroless static
├── docker-compose.yml    # Local dev stack: app + Postgres
├── .dockerignore
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

---

## Quick Start (Docker)

The fastest way to run the service. You need Docker Desktop (or Docker Engine + Compose v2).

```bash
git clone https://github.com/mritunjayupadhyay/go-payroll-api.git
cd go-payroll-api
docker compose up --build
```

That starts two containers:

- `db` — Postgres 16, with `migrations/001_init.sql` auto-applied on first boot
- `app` — the API on `http://localhost:8080`, waits for the DB healthcheck before starting

Sanity check from another terminal:

```bash
curl http://localhost:8080/health
# {"status":"ok"}

curl -X POST http://localhost:8080/api/employees \
  -H 'Content-Type: application/json' \
  -d '{"name":"Jane Doe","hourly_rate":25.00}'

curl -X POST http://localhost:8080/api/payslips/calculate \
  -H 'Content-Type: application/json' \
  -d '{"employee_id":1,"hours_worked":40}'
```

Tear-down:

```bash
docker compose down       # stop containers, keep DB data in the named volume
docker compose down -v    # also wipe the DB volume (next `up` re-applies migrations from scratch)
```

**Image:** the runtime stage is built on `gcr.io/distroless/static-debian12:nonroot`. The final image is ~25 MB and runs as a non-root user with no shell or package manager.

**Note on migrations in compose:** `migrations/` is mounted into Postgres' `/docker-entrypoint-initdb.d`, which runs the SQL files exactly once on a fresh data directory. This is intentionally simple for local dev — a production deployment would use a real migration tool (e.g. `golang-migrate`, `goose`).

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

# Apply the schema (psql lives in the container, so no host install needed)
docker exec -i payroll-db psql -U postgres -d payroll < migrations/001_init.sql

# Run the server (reads DATABASE_URL from the environment)
DATABASE_URL='postgres://postgres:password@localhost:5432/payroll?sslmode=disable' \
  go run ./cmd/server
```

### Running tests

```bash
# Fast suite — handler tests use an in-memory fake store, no DB required.
go test -short ./...

# Full suite — runs the storage integration tests against a real Postgres.
docker exec payroll-db psql -U postgres -c 'CREATE DATABASE payroll_test;'
TEST_DATABASE_URL='postgres://postgres:password@localhost:5432/payroll_test?sslmode=disable' \
  go test ./...

# Coverage
go test -cover ./...
```

The storage package's `TestMain` re-applies `migrations/001_init.sql` on each
run, and each test truncates with `RESTART IDENTITY` for a clean slate.

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

### Create an employee

```bash
curl -X POST http://localhost:8080/api/employees \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "hourly_rate": 25.00}'
```

Response (`201 Created`):
```json
{
  "id": 1,
  "name": "Jane Doe",
  "hourly_rate": 25,
  "created_at": "2026-04-26T02:22:35.906633+07:00"
}
```

### Get an employee

```bash
curl http://localhost:8080/api/employees/1
```

Returns the same shape as `POST /api/employees`. `404` if the employee does not exist.

### Calculate and persist a payslip

```bash
curl -X POST http://localhost:8080/api/payslips/calculate \
  -H "Content-Type: application/json" \
  -d '{"employee_id": 1, "hours_worked": 40}'
```

Response (`201 Created`):
```json
{
  "id": 1,
  "employee_id": 1,
  "hours_worked": 40,
  "gross_pay": 1000,
  "federal_tax": 120,
  "social_security": 62,
  "medicare": 14.5,
  "net_pay": 803.5,
  "created_at": "2026-04-26T02:22:36.004076+07:00"
}
```

The endpoint:
- Looks up the employee (`404` if not found)
- Computes the payslip from the employee's stored `hourly_rate` and the request's `hours_worked`
- Persists the result and returns it with its new ID

Validation errors return `400 {"error": "..."}` for: missing/zero `employee_id`, negative `hours_worked`, or malformed JSON.

### Get a payslip by ID

```bash
curl http://localhost:8080/api/payslips/1
```

Returns the same shape as the calculate response. `404` if not found.

### List payslips for an employee

```bash
curl http://localhost:8080/api/employees/1/payslips
```

Returns an array of payslips for that employee, ordered by `created_at` descending. Always returns `[]` (not `null`) when empty.

### Run the server

```bash
DATABASE_URL='postgres://postgres:password@localhost:5432/payroll?sslmode=disable' \
  go run ./cmd/server
# server listens on :8080
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
- [x] PostgreSQL persistence with migrations and integration tests
- [x] Dockerized deployment (multi-stage build, docker-compose)
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