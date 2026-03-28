# Task 10: Anthropic Messages API Client

Go client for calling the Anthropic Messages API with streaming.

## Steps

- [ ] Create `internal/anthropic/client.go` — wraps the Anthropic Messages API
- [ ] Read API key from `ANTHROPIC_API_KEY` env var; fail at startup if missing
- [ ] Build message payloads: system prompt with optional context (selected text, surrounding text), plus conversation history
- [ ] Support streaming responses (SSE from Anthropic → parsed events)
- [ ] Write tests with a mock HTTP server that returns SSE responses
- [ ] System prompt template: instruct Claude it's helping the user read a research paper, include selected/surrounding text when provided
