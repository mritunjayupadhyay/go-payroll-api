# CLAUDE.md

This file provides project context for Claude Code. Read this first when starting any task in this repo.

---

## Project context

This is a personal learning project to build Go fluency. The author is an experienced senior backend engineer (8+ years) coming from Node.js / NestJS / TypeScript, transitioning into Go. The project is a payroll calculation REST API with PostgreSQL persistence, designed to be portfolio-quality but intentionally scoped — it is not a real payroll system.

The author's goals, in priority order:

1. Learn idiomatic Go — they want to write Go that a Go engineer would write, not Node.js patterns translated into Go syntax.
2. Have a demonstrable project for their CV and interviews within 7 days.
3. Build something that follows industry-standard patterns (layered architecture, dependency injection, testing, containerization).

---

## How to help most effectively

**Explain the "why," not just the "how."** When suggesting code, briefly note why this is the idiomatic Go approach, especially when it differs from Node.js / TypeScript. Example: "In Go, error handling is explicit at every call site rather than try/catch — this is intentional, the language treats errors as values."

**Point out non-idiomatic patterns even if they work.** The author wants to learn, so flag things like:
- Using `interface{}` / `any` when a concrete type would do
- Returning `(result, error)` in the wrong order
- Naming conventions (e.g., `GetUser` not `getUser` for exported, `ID` not `Id`)
- Receiver naming consistency
- Using pointers vs. values inappropriately
- Unnecessary use of channels or goroutines for problems that don't need them

**Prefer standard library over third-party packages** when possible. Go's stdlib is rich. Don't pull in dependencies for things `net/http`, `log/slog`, `database/sql`, or `encoding/json` already do well.

**Show, don't lecture.** When the author asks how to do something, give a working code example with a 2-3 line explanation, not paragraphs of theory.

**Code review is welcome.** When the author says "review this," give honest feedback. They prefer directness over politeness — if something is non-idiomatic, say so plainly.

---

## Stack and conventions

- **Go version:** 1.22+
- **Web framework:** Gin (`github.com/gin-gonic/gin`)
- **Database driver:** `github.com/jackc/pgx/v5/pgxpool`
- **Logging:** `log/slog` (standard library)
- **Testing:** standard `testing` package, `net/http/httptest` for HTTP tests
- **Validation:** Gin's binding tags + manual validation in handlers when needed

### Project layout

Follows the standard Go project layout:

```
cmd/<binary-name>/main.go    # Entry points (cli, server)
internal/<package>/          # Private application code
migrations/                  # SQL migration files
deploy/                      # Dockerfile, docker-compose, k8s manifests
```

Code under `internal/` cannot be imported by other modules — this is intentional for a single-application project.

### Package responsibilities

- `internal/payroll` — Pure calculation logic. **No HTTP, no database imports.** Easy to unit test in isolation.
- `internal/handlers` — HTTP handlers. Thin layer that validates input, calls the calculator and storage, returns responses.
- `internal/storage` — Database access. Exposes a `Storage` interface so handlers depend on the interface, not the concrete implementation. Makes testing easier.

### Naming conventions

- Exported identifiers: `PascalCase` (e.g., `Calculator`, `SavePayslip`)
- Unexported: `camelCase` (e.g., `calculatePayslip`, `taxRates`)
- Acronyms keep their case: `ID`, `URL`, `HTTP` (not `Id`, `Url`, `Http`)
- Receiver names: short, consistent across all methods of a type (e.g., `c *Calculator`, `s *Storage`)
- Test files: `_test.go` suffix, same package as the code being tested

### Error handling

- Always return errors, never panic in production paths. Panics are for truly unrecoverable situations (e.g., invalid configuration at startup).
- Wrap errors with context using `fmt.Errorf("doing X: %w", err)` so the call chain is visible.
- Don't ignore errors with `_` unless there's a clear reason; document why if you do.
- Sentinel errors (e.g., `var ErrNotFound = errors.New("not found")`) are fine for expected error cases the caller might want to check with `errors.Is`.

### Testing

- Prefer table-driven tests for functions with multiple input/output cases.
- Use `httptest.NewRecorder()` and `httptest.NewRequest()` for HTTP handler tests — no need to spin up a real server.
- Don't mock what you can fake. A simple in-memory storage implementation is often better than a mock library.
- Test the public API of each package, not internal helpers.

---

## What this project is NOT

- It is not a real payroll system. Tax calculations are mock and intentionally simplified.
- It is not a microservices showcase. It's a single service, deliberately. Adding microservices for a learning project is over-engineering.
- It is not trying to demonstrate every Go feature. Goroutines, channels, generics — only use them where they're genuinely the right tool. Forced demonstrations look amateurish in a portfolio.
- It is not production-grade. There's no real auth, no audit log, no compliance handling. Adding these would be scope creep.

---

## Milestone tracking

The project is built in milestones. Each milestone produces a working, demonstrable state. When the author asks for help, check which milestone they're on:

- **M0:** Setup — Go module initialized, hello world works.
- **M1:** CLI — Core calculation logic with unit tests, runnable as a CLI.
- **M2:** REST API — HTTP endpoints using Gin, input validation, integration tests.
- **M3:** Persistence — PostgreSQL storage, migrations, CRUD endpoints.
- **M4:** Docker — Multi-stage Dockerfile, docker-compose for local dev.
- **M5:** Polish — README, structured logging, CI workflow, release tag.
- **M6 (optional):** Kubernetes — kind cluster, manifests, screenshots.

Don't suggest features from a later milestone unless the author asks. If a fix or improvement happens to be from a later milestone but is small, mention it briefly: "FYI, this will become relevant in M3 when we add the database — happy to flag patterns now or wait."

---

## Common requests and how to handle them

**"Is this idiomatic?"** — Be honest. Point out anything that isn't, and show the idiomatic version. Brief explanation of why.

**"Why does Go do X this way?"** — Give a short, concrete answer. Reference the Go proverbs or design rationale when relevant. Example: "Channels orchestrate, mutexes serialize" is a useful framing.

**"Compare this to Node.js / NestJS"** — These comparisons help the author build mental bridges. Use them when the parallel is genuinely informative, not as filler.

**"Refactor this"** — Refactor minimally. Don't restructure the whole project unless asked. Explain each change in 1-2 lines.

**"Add feature X"** — Check whether X is in the current milestone. If not, ask whether to add it now or note it for later. Don't expand scope silently.

---

## Things to avoid

- Don't write Java-style Go (excessive abstraction, factory patterns, deep inheritance simulation).
- Don't over-use goroutines and channels. Most code in this project is synchronous and that's correct.
- Don't add ORMs (GORM, etc.) — the author is learning Go's database/sql and pgx idioms directly.
- Don't add unnecessary middleware layers. Gin's built-in middleware covers logging and recovery; we'll add request ID middleware in M5 specifically.
- Don't suggest premature optimization. Clarity first, performance later if measurements justify it.

---

## Author preferences

- Concise responses. Code with brief explanation beats long explanation with no code.
- When uncertain about the right approach, present 2 options with tradeoffs rather than picking arbitrarily.
- README and code comments should be honest — this is a learning project, and that framing is fine to acknowledge.


## Build plan

The full milestone-by-milestone plan lives in `BUILD_PLAN.md` in the repo root.
Before suggesting changes or features, check which milestone the user is currently
working on. Don't suggest features from later milestones unless explicitly asked.
If a small improvement is technically from a later milestone, mention it briefly
rather than implementing it silently.