# Task 55: Decouple chat API from Anthropic types with intermediate representation

Introduce domain-level message and content types in the API layer so that Anthropic-specific structures don't leak throughout the project. This is a planning/design task — implementation in a follow-up.

## Context

Currently `internal/api/messages.go` (639 lines) is deeply coupled to `internal/anthropic` types. The `ChatStreamer` interface, tool loop, history reconstruction, and content block persistence all operate directly on `anthropic.Message`, `anthropic.ContentBlock`, `anthropic.StreamEvent`, etc. This makes it impossible to swap providers and makes the handler hard to reason about.

### Where Anthropic types leak today

**ChatStreamer interface** (`internal/api/messages.go:24-26`):
- `Stream(ctx, anthropic.Request) (<-chan anthropic.StreamEvent, error)` — any implementation must use Anthropic types

**Store → API conversion** (`messages.go:150-193`):
- Deserializes `store.Message.ContentBlocks` JSON into `[]anthropic.ContentBlock`
- Builds `anthropic.Message` with content blocks and image sources

**Tool loop** (`messages.go:244-358`):
- Collects `anthropic.StreamEvent` tool calls
- Builds `anthropic.ContentBlock` for tool_use and tool_result
- Builds `anthropic.Message` for conversation turns

**Content block helpers** (`messages.go:383-472`):
- `persistContentBlocks()` marshals `[]anthropic.ContentBlock` to JSON
- `buildMultimodalUserMessage()` returns `anthropic.Message`
- `buildBlocksFromPersistedAttachments()` returns `[]anthropic.ContentBlock`

**Tool executor** (`messages.go:517-616`):
- `toolExecResult` contains `[]anthropic.ContentPart` and `anthropic.ImageSource`

**Main server** (`cmd/research-server/main.go`):
- Constructs `anthropic.Client` directly

### What's already clean

- **Store layer**: Only knows `string` content and `*string` JSON — no anthropic imports
- **Frontend**: Generic `ToolCall`/`ToolResult` interfaces over SSE JSON — no coupling
- **SSE types**: `sseResponse`, `sseToolCall`, `sseToolResult` are already anthropic-free

## Design questions to resolve

- [ ] Where should the intermediate types live? Options: `internal/chat/` package, or types in `internal/api/`
- [ ] Should `ChatStreamer` accept/return domain types, with the Anthropic client wrapped in an adapter? Or should conversion happen at the handler boundary?
- [ ] Should `content_blocks` JSON in the DB store domain types or remain Anthropic-shaped? (Migration concern for existing data)
- [ ] Should tool definitions (`PDFTools()`) be expressed in domain types and converted per-provider, or stay provider-specific?
- [ ] How to handle provider-specific features (e.g. Anthropic's `system` field vs OpenAI's system message convention)?

## Proposed direction (to validate)

1. **Domain types** in a new `internal/chat/` package:
   - `chat.Message` with `Role`, `Content`, `ContentBlocks []chat.ContentBlock`
   - `chat.ContentBlock` covering text, tool_use, tool_result, image
   - `chat.Request` with messages, system prompt, tools
   - `chat.StreamEvent` for deltas, tool calls, stop

2. **Provider interface**:
   - `chat.Provider` with `Stream(ctx, chat.Request) (<-chan chat.StreamEvent, error)`
   - Anthropic adapter converts domain types ↔ `anthropic.*` types

3. **Handler refactor**:
   - `messages.go` operates on `chat.*` types only
   - Conversion to/from store stays in the handler (or a thin adapter)
   - Tool executor returns domain types, not `anthropic.ContentPart`

4. **DB compatibility**:
   - Domain `ContentBlock` JSON should be a superset of what's stored today
   - Old data without `content_blocks` continues to work (already the case)

## Notes

- This task is planning only. The implementation should be a separate task once the design is agreed upon.
- Related to task 47 (LLM provider abstraction) — this task focuses on the intermediate representation; task 47 focuses on OpenAI-compatible alternatives. They share the provider interface concern.
- The `internal/anthropic/` package itself doesn't need to change — it's the leakage into `internal/api/` that's the problem.
