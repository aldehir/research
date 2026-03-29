# Task 46: Custom system prompt for research/textbook assistant

Construct a richer system prompt that goes beyond the current minimal instructions. The prompt should establish the assistant as an explanatory research and textbook companion — answering in a clear, pedagogical tone with structured explanations rather than terse replies.

## Context

The current system prompt is a single sentence in `internal/anthropic/prompt.go` (`basePrompt` constant). Document metadata (title, author, date, page count) is appended by `BuildSystemPromptFromContext()`. The `Client.Stream()` method in `client.go` (line 208) falls back to this builder when `req.SystemPrompt` is empty — so the existing plumbing already supports an explicit system prompt override, but nothing populates it.

Key files:
- `internal/anthropic/prompt.go` — `basePrompt`, `PromptContext`, `BuildSystemPromptFromContext()`
- `internal/anthropic/client.go` — `Request.SystemPrompt`, `Client.Stream()` fallback logic
- `internal/api/messages.go` — handler that builds the `anthropic.Request`

No frontend changes needed — this is purely backend prompt engineering.

## Checklist

- [x] Write test for `BuildSystemPromptFromContext` asserting the new prompt contains key behavioral instructions (explanatory tone, structured answers, citation of page/section)
- [x] Rewrite `basePrompt` with detailed persona and behavioral guidelines (explanatory tone, step-by-step reasoning, cite evidence from the document, handle ambiguity gracefully)
- [x] Ensure document metadata block (title/author/date/pages) is still appended correctly
- [x] Verify existing prompt tests still pass after rewrite
- [ ] Manual smoke test: ask the assistant a conceptual question and confirm the response is explanatory rather than terse

## Notes

- The prompt should instruct the model to:
  - Act as a knowledgeable research and textbook reading companion
  - Answer in an explanatory, pedagogical tone — like a patient tutor
  - Provide structured explanations (definitions first, then context, then implications)
  - Reference specific pages/sections from the document when possible
  - When uncertain, say so rather than fabricate
  - Adapt depth to the complexity of the question
- Keep the prompt concise enough to avoid eating too many tokens from the context window
- Do not add any user-facing configuration for the prompt — this is a backend-only change
