# Task 02: Add comprehensive logging with log/slog

Add structured logging to the Go backend using the stdlib `log/slog` package, replacing ad-hoc `log` and `fmt` calls. Introduce request logging middleware, and instrument handlers, store, anthropic client, and PDF storage.

## Context

- **Current state**: Only 5 log statements exist, all in `cmd/server/main.go` using `log.Printf`/`log.Fatalf` and one `fmt.Printf`. No logging in handlers, store, anthropic client, or PDF storage.
- **No middleware**: `internal/api/api.go` registers handlers directly on `http.ServeMux` — no middleware chain exists.
- **Handler pattern**: Closures returning `http.HandlerFunc` with dependencies injected (e.g. `handleListPapers(db *sql.DB) http.HandlerFunc`). Error responses go through `writeError(w, status, msg)`.
- **Packages to instrument**:
  - `cmd/server/main.go` — startup/shutdown
  - `internal/api/` — request logging middleware, error logging in handlers
  - `internal/store/` — database operation errors
  - `internal/anthropic/` — API call logging (request start, stream completion, errors)
  - `internal/pdf/` — file I/O operations
- **Dependencies**: None (uses stdlib `log/slog`)

## Checklist

- [x] Initialize `slog.Logger` in `main.go`, replace existing `log`/`fmt` calls with slog equivalents
- [x] Add request logging middleware in `internal/api/` (method, path, status, duration)
- [x] Wire middleware into `NewMux` and add `*slog.Logger` parameter
- [x] Add slog logging to handler error paths (via `writeError` or individual handlers)
- [x] Add slog logging to `internal/store/` operations (errors, key operations)
- [x] Add slog logging to `internal/anthropic/` client (stream start, completion, errors)
- [x] Add slog logging to `internal/pdf/` storage (save, delete, errors)
- [x] Update existing tests to account for logger parameter changes

## Notes

- Use `slog.Default()` or pass `*slog.Logger` explicitly — prefer explicit injection where the dependency is already threaded (api, store) to keep testability.
- For request logging middleware, capture status code with a `responseWriter` wrapper.
- Consider adding a request ID to the logger context for correlation.
- Keep log levels sensible: `Info` for request log lines, `Error` for failures, `Debug` for verbose detail like Anthropic streaming events.
