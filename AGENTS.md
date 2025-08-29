# Agent Guide (Lighthouse)

1. Build binaries: `go build ./...`; run services: `make run-orchestrator`, `make run-host-agent`, `make run-registry-monitor`.
2. Generate protos: `make protos`; SQL codegen: `make sqlc/orchestrator`.
3. DB setup: `make make-db` (runs postgres, createdb, migrate-up). Tear down: stop container manually, `make dropdb` if needed.
4. Tests: `go test ./...`; single test file: `go test path/to/pkg -run TestName`; with coverage: `go test -cover ./...`.
5. Lint/format: `go fmt ./...`; enforce imports with `goimports` (group: std, third-party, internal `github.com/MadhavKrishanGoswami/Lighthouse/...`). No unused imports.
6. Go version: 1.24.x (see go.mod). Use generics conservatively; prefer simple explicit types.
7. Naming: exported identifiers CamelCase; unexported lowerCamel. Constants ALL_CAPS or CamelCase if typed. gRPC / proto fields keep generated naming.
8. Errors: never ignore errors; wrap with context `fmt.Errorf("<action>: %w", err)`; log once at boundary (main or server handlers). Return sentinel or typed errors for branch logic.
9. Logging: use `log.Printf`; no panics except truly unrecoverable (e.g., cannot start server). `log.Fatalf` only in `main` startup failures.
10. Concurrency: guard shared maps with RWMutex (pattern shown in agent server). Always respect context cancellation; pass ctx downward.
11. Streams: for gRPC bidirectional streams, separate read/write loops; close done channels; avoid blocking by buffering command channels.
12. Database: use sqlc generated methods; do not write raw SQL in app code; extend queries via `services/orchestrator/db/queries` then regenerate.
13. Migrations: modify new sequential file pairs; never edit applied migrations; run `make migrate-up`.
14. Proto changes: edit `.proto`, run `make protos`, commit generated code (adjust .gitignore if choosing to commit) else ensure generation in CI.
15. Import ordering: std lib, blank line, external deps, blank line, internal module paths.
16. Timeouts: use `context.WithTimeout` for external calls (registration heartbeat already uses 10s); avoid magic numbers—extract constants.
17. Testing: create `_test.go` alongside code; use table-driven tests; prefer `t.Helper()` for helpers; avoid global state.
18. Configuration: use `cleanenv`; provide example defaults in `local.example.yaml`; never commit secrets; `local.yaml` is ignored.
19. Generated artifacts (`*.pb.go`, sqlc) treated as read-only; do not hand-edit.
20. If adding tooling (lint, CI), keep commands minimal and document here; currently no Cursor/Copilot rule files present.
