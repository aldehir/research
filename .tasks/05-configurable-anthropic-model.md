# Task 05: Make Anthropic model configurable via environment variable

Fix 400 error from Anthropic API by making the model configurable. The model is currently hardcoded to `claude-sonnet-4-20250514` in the client, which may be invalid/deprecated. Adding an env var lets the user override it.

## Context

- `internal/anthropic/client.go`: `Client` struct has a `Model` field, hardcoded in `NewClient()` to `claude-sonnet-4-20250514` (line 46)
- `cmd/server/main.go`: creates client via `anthropic.NewClient(apiKey)` — only passes API key
- Error surfaces as `anthropic api: status 400` from `client.go:117` (non-200 status check)
- Error body from Anthropic is discarded — not logged, making diagnosis harder

## Checklist

- [x] Read and log the Anthropic error response body on non-200 status
- [x] Add `ANTHROPIC_MODEL` env var support with sensible default
- [x] Pass model from `main.go` into the client (e.g. `NewClient` option or setter)
- [x] Verify default model ID is valid (update if needed)

## Notes

- The current error handler discards the response body — logging it would immediately reveal whether the 400 is a model issue or something else
- Default model should be updated to a known-valid ID (e.g. `claude-sonnet-4-20250514` or latest)
- Consider whether the `.env` file should get an entry too
