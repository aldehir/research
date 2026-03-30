# Task 61: Background support for chat streaming

Allow LLM responses to complete in the background even when the user navigates away from the chat. Currently, navigating away aborts the frontend fetch, which cancels the Go HTTP request context, killing the Anthropic API stream mid-response.

## Context

**The cancellation chain today:**
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

- [x] Backend: decouple streaming from HTTP request context — use a detached context for the Anthropic API call in `handleSendMessage` so the LLM stream and tool loop continue even if the client disconnects
- [x] Backend: buffer and persist assistant response chunks to the database as they arrive, so a completed-in-background response is available when the user returns
- [x] Backend: add a lightweight status/reconnect endpoint (e.g. `GET /api/papers/{id}/chats/{id}/stream`) that lets the frontend poll or reconnect to an in-progress stream, receiving any buffered events it missed
- [x] Frontend: stop aborting the backend stream on navigation — either don't call `abortActiveStream()` at all, or only abort the local reader without closing the HTTP connection (the backend will finish independently regardless)
- [x] Frontend: on (re-)entering a chat that has an in-progress or recently-completed background stream, reconnect to the status endpoint and replay buffered events so the UI catches up
- [x] Add tests for background completion: verify that cancelling the HTTP request context does not stop the Anthropic stream from finishing and being persisted
- [x] Add tests for reconnect: verify the frontend receives the full response when returning to a chat whose stream completed in the background

## Notes

- The tool execution loop in `messages.go` (lines 239-355) can make multiple sequential Anthropic API calls (for tool use). The background context must survive the entire loop, not just one call.
- Need to decide on cleanup: how long to keep in-flight stream state in memory. A reasonable approach is to persist chunks to the DB as they arrive and drop in-memory state once the stream completes.
- The frontend `isStale()` check in `chat.svelte.ts` (line 142) already guards against cross-chat contamination — this logic should be preserved or adapted for the reconnect flow.
- Consider whether the backend should also handle the case where the *server* restarts mid-stream (out of scope for this task, but worth noting).
