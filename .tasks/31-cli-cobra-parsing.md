# Task 31: Add CLI parsing with cobra, rename binary to research-server

Add cobra for CLI argument parsing to the server binary. Rename the entrypoint from `cmd/server` to `cmd/research-server` so `go install` produces a `research-server` binary.

## Context

- Single entrypoint at `cmd/server/main.go` — all config is via env vars (`ADDR`, `DB_PATH`, `PDF_DIR`, `LOG_LEVEL`, `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`, `FRONTEND_DIR`)
- No CLI flags or subcommands exist yet
- cobra is not in `go.mod` — needs `go get github.com/spf13/cobra`
- Module: `github.com/aldehir/research`
- `frontend.BuildFS` is embedded via `frontend/` package
- `runIndexer`, `resolveFrontendFS`, `serveFrontend` are helper functions in `main.go`

## Checklist

- [x] Rename `cmd/server/` to `cmd/research-server/`
- [x] Add `github.com/spf13/cobra` dependency
- [x] Create root cobra command with `serve` as default action
- [x] Convert env-var config to cobra flags with env-var fallbacks (`--addr`, `--db-path`, `--pdf-dir`, `--log-level`, `--frontend-dir`)
- [x] Keep `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL` as env-only (secrets shouldn't be CLI flags)
- [x] Verify `go build ./cmd/research-server` produces working binary
- [x] Update CLAUDE.md run instructions if needed

## Notes

- cobra flags should use env vars as defaults (e.g. `os.Getenv` in flag default) so existing env-var-based deployments keep working
- Consider whether subcommands are needed now or just a root command — start with root only unless the user wants subcommands
