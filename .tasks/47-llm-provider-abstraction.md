# Task 47: Abstract LLM provider with OpenAI-compatible alternative

Extract a provider-agnostic LLM interface from the current Anthropic implementation, then add an OpenAI Chat Completions API provider. Both providers must support tool calling, vision, and streaming (SSE).

## Context

Summary of relevant existing code discovered during exploration:

- **`internal/anthropic/client.go`** — Anthropic API client. Key types: `Client`, `Request`, `Message`, `ContentBlock`, `ContentPart`, `ImageSource`, `StreamEvent`. The `Stream()` method returns `<-chan StreamEvent` and handles SSE parsing from Anthropic's wire format. Custom JSON marshal/unmarshal on `Message` and `ContentBlock` handles both plain text and structured (tool_use/tool_result) content.
- **`internal/anthropic/tools.go`** — `Tool` struct and `PDFTools()` returning tool definitions with JSON schemas. The `Tool` type (Name, Description, InputSchema) is already provider-agnostic in shape.
- **`internal/anthropic/prompt.go`** — `BuildSystemPromptFromContext()` builds a system prompt string. Provider-agnostic logic.
- **`internal/api/messages.go`** — `ChatStreamer` interface (`Stream(ctx, anthropic.Request) (<-chan anthropic.StreamEvent, error)`) is the existing abstraction point. `handleSendMessage()` runs the tool execution loop consuming `StreamEvent` channels. Tool execution (`executeToolCall`) is already provider-agnostic. SSE events sent to the frontend (delta, tool_call, tool_result, done, error) are also provider-agnostic.
- **`internal/api/api.go`** — Route registration, takes `ChatStreamer` dependency.
- **`cmd/research-server/main.go`** — Initializes `anthropic.Client` from env vars, passes as `ChatStreamer` to `api.NewMux()`.
- **Frontend** (`api.ts`, `chat.svelte.ts`) — SSE consumption is already provider-agnostic; no changes needed.

### Current coupling to Anthropic

1. `ChatStreamer` interface references `anthropic.Request` and `anthropic.StreamEvent`
2. `handleSendMessage()` builds `anthropic.Message`, `anthropic.ContentBlock`, `anthropic.ContentPart` directly
3. `toolExecResult` contains `anthropic.ContentPart` for image results
4. `anthropic.PDFTools()` returns `anthropic.Tool`
5. `main.go` creates `anthropic.Client` directly

### Design approach

Create `internal/llm/` package with provider-agnostic types and interface:
- Move types (`Message`, `ContentBlock`, `ContentPart`, `ImageSource`, `StreamEvent`, `Tool`, `Request`) to `internal/llm/`
- Define `Streamer` interface in `internal/llm/` (replaces `ChatStreamer`)
- Refactor `internal/anthropic/` to implement `llm.Streamer` using `internal/llm/` types
- Create `internal/openai/` implementing `llm.Streamer` via OpenAI Chat Completions API
- Move prompt logic to `internal/llm/` (it's already provider-agnostic)
- Move tool definitions to `internal/llm/` (already provider-agnostic)
- Update `internal/api/` to use `llm.*` types instead of `anthropic.*`
- Update `cmd/research-server/main.go` to select provider from config/env

## Checklist

- [ ] Create `internal/llm/` package with provider-agnostic types (Message, ContentBlock, ContentPart, ImageSource, StreamEvent, Tool, Request)
- [ ] Define `Streamer` interface in `internal/llm/` and move prompt/tool logic there
- [ ] Refactor `internal/anthropic/` to translate between `llm.*` types and Anthropic wire format
- [ ] Update `internal/api/messages.go` to use `llm.*` types instead of `anthropic.*`
- [ ] Update `internal/api/api.go` and `cmd/research-server/main.go` to use `llm.Streamer`
- [ ] Verify existing tests pass with refactored types
- [ ] Implement `internal/openai/` client: OpenAI Chat Completions API with streaming, tool calling, and vision support
- [ ] Add tests for OpenAI client (SSE parsing, tool call handling, image content)
- [ ] Add provider selection via env var (`LLM_PROVIDER=anthropic|openai`) in main.go
- [ ] End-to-end manual verification with both providers

## Notes

- The OpenAI Chat Completions API uses a different SSE format (chunked `choices[0].delta`) and different tool_call structure (`tool_calls` array in delta with index-based accumulation). The OpenAI client must translate this to the common `StreamEvent` channel.
- Vision in OpenAI uses `content: [{type: "image_url", image_url: {url: "data:image/png;base64,..."}}]` rather than Anthropic's `source` block. The provider must translate `ImageSource` to the appropriate wire format.
- Tool definitions use `"parameters"` in OpenAI vs `"input_schema"` in Anthropic — translation needed at serialization time.
- The `Tool` type shape (name, description, JSON schema) is essentially the same across providers; only the wire field name differs.
- System prompt is a top-level `system` field in Anthropic but a `role: "system"` message in OpenAI — provider handles this.
- `max_tokens` is required in Anthropic but optional in some OpenAI models — provider should set sensible defaults.
