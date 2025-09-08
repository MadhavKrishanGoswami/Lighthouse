# Agent Quick Guide (Lighthouse)
1. Build services: `make run-host-agent`, `make run-orchestrator`, `make run-registry-monitor`, `make run-tui` (each target builds then runs). For build-only use underlying `go build -o ./services/<svc>/build ./services/<svc>/cmd/.../main.go`.
2. Proto generation: `make protos` (writes to services/common/genproto/* with source_relative paths). Re-run after editing any *.proto.
3. Database: `make postgres` (launch), `make createdb`, `make migrate-up` / `make migrate-down`, or `make make-db` for full setup. Connection URL: postgres://dev:dev@localhost:5432/lighthouse?sslmode=disable.
4. Tests: none present yet. Add *_test.go files and run `go test ./...`. Single test: `go test ./path/to/pkg -run ^TestName$ -count=1`.
5. Lint/format: use `go vet ./...` and `gofmt -s -w .` (enforce before commits). Prefer `staticcheck` if available.
6. Imports: stdlib first, blank line, third-party, blank line, internal (`github.com/MadhavKrishanGoswami/Lighthouse/...`). No unused imports; keep deterministic ordering via `gofmt`.
7. Naming: camelCase for locals, PascalCase for exported, ALL_CAPS avoided (use const with PascalCase). Filenames lowercase underscore only if needed.
8. Errors: return `error` as last value; wrap context with `fmt.Errorf("context: %w", err)`. Do not panic for recoverable conditions. Log once at boundary (e.g., server handlers); inner functions just propagate.
9. Logging: (No logger abstraction yet) prefer `log` or future structured logger; avoid fmt.Println outside prototypes.
10. Concurrency: prefer context-aware operations; pass `context.Context` as first param when adding new I/O or RPC funcs.
11. gRPC / Protobuf: Regenerate code after proto change; never edit generated files. Keep service-specific packages segregated as current structure.
12. SQL (sqlc): Edit .sql in queries dir, then `make sqlc/orchestrator`; never modify generated code manually.
13. Configuration: Use cleanenv patterns (see config packages). Keep env var names consistent; validate required fields.
14. Modules & deps: manage with `go get`, prune with `go mod tidy` after changes.
15. TUI: UI components under services/tui/internal/ui; keep presentation (rendering) separate from data client layer.
16. Avoid circular dependencies: only internal packages depend downward (e.g., orchestrator/monitor -> orchestrator/db/sqlc, not vice versa).
17. Add tests for DB queries (requires running postgres + migrations). Use transactions with rollback in tests where possible.
18. Performance: prefer streaming (e.g., updateStream) over polling; reuse gRPC clients instead of recreating.
19. Secrets: none committed; keep future secrets in env / local.yaml; NEVER commit credentials.
20. Housekeeping: run format + vet + tests + protos (if changed) + sqlc (if SQL changed) before commit.
