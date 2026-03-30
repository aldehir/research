# Task 61: Background support for chat streaming

Allow LLM responses to complete in the background even when the user navigates away from the chat. Previously, navigating away aborted the frontend fetch, which cancelled the Go HTTP request context, killing the Anthropic API stream mid-response.

## Context

**The cancellation chain before this task:**
1. Frontend `chat.svelte.ts` calls `abortActiveStream()` on navigation (new chat, switch chat, reset)
2. `AbortController.abort()` cancels the fetch in `api.ts:sendMessage()`
3. Client closes HTTP connection → Go's `r.Context()` is cancelled
4. `messages.go:handleSendMessage` passes `r.Context()` to `provider.Stream()`
5. `anthropic/client.go:Stream()` creates HTTP request with that context → cancelled
6. `readSSE()` checks `ctx.Done()` per line → exits, closing channel
7. Tool loop in `messages.go` exits, response is lost

**Key files:**
- `frontend/src/lib/chat.svelte.ts` — AbortController management, `abortActiveStream()`
- `frontend/src/lib/api.ts` — `sendMessage()` with fetch + SSE parsing
- `internal/api/messages.go` — `handleSendMessage`, SSE writer, tool execution loop
- `internal/anthropic/client.go` — `Stream()`, `readSSE()` with context cancellation
- `internal/anthropic/adapter.go` — context passthrough to Anthropic client
- `internal/chat/provider.go` — `Provider` interface with `Stream(ctx, req)`
- `internal/store/` — SQLite persistence for messages

## Checklist

- [x] Backend: use `context.Background()` for provider stream so the handler goroutine survives client disconnect — writes to ResponseWriter silently fail, final assistant message is persisted to DB at the end
- [x] Backend: add `RunningStreams` guard to prevent duplicate concurrent streams to the same chat (409 Conflict)
- [x] Frontend: replace `abortActiveStream()` with `detachStream()` — clears local UI state without aborting the fetch, so the backend continues processing
- [x] Add tests for background completion: verify that cancelling the HTTP request context does not stop the stream from finishing and being persisted
- [x] Add tests for conflict detection: verify 409 when a stream is already running

## Notes

- The handler stays synchronous — no background goroutines, no event buffers, no reconnect endpoint. The Go HTTP server does not kill handler goroutines on client disconnect; `r.Context()` is cancelled but the goroutine continues.
- When the user navigates back to a chat whose stream completed in the background, `selectSession` loads the persisted messages from the DB.
