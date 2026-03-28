# Task 11: Add detailed logging for LLM messages and tool calling

Add structured logging to introspect the full LLM request/response lifecycle: messages sent, tool calls made, tool results returned, and iteration counts.

## Context

Summary of relevant existing code:
- **`internal/anthropic/client.go`**: `Stream()` builds `apiRequest` and POSTs to Anthropic. Currently logs stream start (model, message count) and errors only. Does not log request body, system prompt, tool definitions, or response content.
- **`internal/api/messages.go`**: `handleSendMessage()` runs the tool loop (up to 10 iterations). Currently logs stream start failure and tool execution warnings. Does not log: conversation history sent, tool call details, tool results, iteration count, or total response.
- **`internal/anthropic/tools.go`**: Defines 3 PDF tools (`search_pdf`, `read_page`, `go_to_page`). No logging on tool definitions sent.
- All logging uses `log/slog` with structured key-value pairs.

## Checklist

- [x] Add `LOG_LEVEL` env var to `cmd/server/main.go` to enable debug output (e.g. `LOG_LEVEL=DEBUG`)
- [x] Log full request details in `client.go` Stream: system prompt length, tool count, message roles/types, and request payload at Debug level
- [x] Log each SSE event type received during streaming at Debug level
- [x] Log tool loop iteration count and tool call details (name, args) in `messages.go`
- [x] Log tool execution results (name, result length, duration) in `messages.go`
- [x] Log final assistant response summary (length, total tool iterations, total duration)
- [x] Add tests verifying log output contains expected structured fields

## Notes

- Use `slog.Debug` for verbose message content to avoid noise at default log levels
- Use `slog.Info` for high-level summaries (iteration count, total duration, tool names used)
- Truncate long values (page text, search results) in logs to keep output readable
- Consider a `slog.Group` for related fields (e.g., `slog.Group("tool", "name", name, "duration", d)`)
