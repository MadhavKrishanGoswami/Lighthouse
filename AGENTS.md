# Agent Guide (Lighthouse)

Build & Run: `make protos`; `go build ./...`; services via `make run-host-agent`, `make run-orchestrator`, `make run-registry-monitor`. Generate SQL: `make sqlc/orchestrator`. DB setup: `make make-db` (runs postgres, createdb, migrate-up). Migrations: `make migrate-up|migrate-down`.
Tests: No test files yet; add *_test.go near code. Run all: `go test ./...`. Single package: `go test ./services/orchestrator/internal/grpc/agent -v`. Single test: `go test -run TestName -count=1 ./path/to/pkg`. Race: add `-race`. Coverage: `go test -cover ./...`.
Lint/Format: Use `go vet ./...`, `golangci-lint run` (if added), `go fmt ./...` before commit. Keep imports grouped: std, third-party, internal (github.com/MadhavKrishanGoswami/Lighthouse/...). No unused imports.
Proto: Source in `proto/`, regenerate only via Make target; do not edit generated `services/common/genproto/**` manually.
SQLC: Edit `.sql` in `services/orchestrator/db/queries/`, regenerate with make target; never edit generated `internal/db/sqlc/*.go` directly.
Errors: Return `(val, err)` early; wrap with context using `fmt.Errorf("action: %w", err)` (avoid log+return duplication). Log only at edges (gRPC handlers, main). Prefer `errors.Is/As` over string match.
Logging: Use `log.Printf` (current code) with concise, contextual messages; avoid sensitive data.
Naming: Go conventions: exported CamelCase, unexported camelCase, acronyms (ID, IP, gRPC -> gRPC in names). Packages: short, lower case. File headers optional; keep package comments meaningful.
Concurrency: Protect shared maps with `sync.Mutex` (as in Server.Hosts). Do not hold locks across network calls.
Types: Prefer concrete types; use slices rather than pointers to slices unless nil distinction needed. Use `context.Context` as first param when passing.
Configuration: Use `cleanenv`; add new fields to `config.go` and sample `local.example.yaml`; never commit filled `local.yaml`.
Git Hygiene: Do not commit generated binaries (`build`), `go.sum` currently ignored (consider committing for reproducibility). Keep diffs minimal.
Add tests: prefer table-driven tests; for DB, isolate using test schema or transaction rollback.
Security: Validate external inputs (future work). Avoid panics in library code.
If adding tools, update this file succinctly; keep <25 lines total.
