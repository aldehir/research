# Task 48: Lua interpreter with eval endpoint

Add a sandboxed Lua interpreter (gopher-lua) and a `POST /api/lua/eval` endpoint. Lua code fences in chat messages get a "Run" button that sends code to the endpoint and displays output inline. Update the system prompt to mention Lua availability.

## Context

- **Backend API**: `internal/api/api.go` registers routes on `http.ServeMux` via `NewMux()`. Handlers follow `handleXxx(deps) http.HandlerFunc` pattern, wrapped with `requestLogger`. JSON helpers: `writeJSON()`, `writeError()`.
- **Markdown rendering**: `frontend/src/lib/MarkdownRenderer.svelte` uses `marked` + `highlight.js` + `DOMPurify`. Code blocks already get syntax highlighting and a copy button overlay. Lua code fences will need an additional "Run" button.
- **System prompt**: `internal/anthropic/prompt.go` — `BuildSystemPromptFromContext()` builds the prompt. Add a note about Lua interpreter availability.
- **Chat streaming**: `internal/api/messages.go` handles SSE streaming and tool execution. Not directly involved here since Lua eval is a separate endpoint, not a tool call.
- **Go deps**: `go.mod` — will need `github.com/yuin/gopher-lua`.

## Design

**Why not a tool call?** Tool calls render as collapsible chips in `MessageThread.svelte`. Lua code should stay visible inline as a code block. Instead, the model naturally writes ` ```lua ` fences, and the frontend adds a play button — no special streaming/tool plumbing needed.

**Sandboxing**: gopher-lua runs in a Go goroutine with no OS access. Remove dangerous stdlib modules (`os`, `io`, `loadfile`, `dofile`). Set execution timeout (5s) and memory limits.

**Frontend flow**: User clicks "Run" on a Lua code block → `POST /api/lua/eval` with `{code: "..."}` → response `{output: "...", error: "..."}` → display result below the code block.

## Checklist

- [x] Add gopher-lua dependency
- [x] Create `internal/lua/` package with sandboxed evaluator (timeout, no OS/IO libs)
- [x] Write tests for Lua evaluator (success, errors, timeout, sandbox restrictions)
- [x] Add `POST /api/lua/eval` endpoint in `internal/api/`
- [x] Write tests for eval endpoint (valid code, syntax error, timeout, empty input)
- [x] Register endpoint in `NewMux()`
- [x] Add `evalLua()` API function in `frontend/src/lib/api.ts`
- [x] Add "Run" button to Lua code blocks in `MarkdownRenderer.svelte`
- [x] Display eval output/error below code block after execution
- [x] Update system prompt in `internal/anthropic/prompt.go` to mention Lua availability
- [x] End-to-end manual test: send message requesting Lua code, run it, verify output

## Notes

- gopher-lua is pure Go (no CGO), consistent with the project's SQLite choice
- Execution timeout should be configurable but default to 5 seconds
- Consider a `print()` override that captures output to a buffer (gopher-lua's default print goes to stdout)
- highlight.js already supports Lua syntax highlighting
- The Run button should show a loading spinner while eval is in progress, then display output in a result block below the code fence
