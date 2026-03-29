# Task 54: Allow overwriting the Anthropic API URL

Allow configuring the Anthropic API base URL via environment variable so a MITM proxy (e.g. mitmproxy, Charles) can be attached for debugging HTTP traffic.

## Context

- `internal/anthropic/client.go`: `Client` struct has a `BaseURL` field, hardcoded to `"https://api.anthropic.com"` in `NewClient()`. Used as `c.BaseURL + "/v1/messages"` when building requests.
- The client already uses an **Option pattern** (`type Option func(*Client)`) with `WithModel()` as the existing example.
- `cmd/research-server/main.go`: Anthropic settings are env-only (no CLI flags) — `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL` are read from `os.Getenv` inside `runServe()`.
- `cmd/research-server/main_test.go`: Has a test asserting Anthropic config stays env-only (not exposed as CLI flags).

## Checklist

- [x] Add `WithBaseURL(url string) Option` to `internal/anthropic/client.go` (test that it overrides the default)
- [x] Read `ANTHROPIC_BASE_URL` env var in `cmd/research-server/main.go` and pass `WithBaseURL` option to `NewClient`
- [x] Log the configured base URL at startup so proxy attachment is observable

## Notes

- Follow the same env-only pattern as `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL` — no CLI flag.
- Strip trailing slash from the URL if provided to avoid double-slash in the path.
