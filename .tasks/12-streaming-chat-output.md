# Task 12: Add streaming output to chat panel

Show the full assistant response lifecycle in the chat panel: streaming text deltas, tool call invocations with expandable results, and the final assembled response.

## Context

Summary of relevant existing code:

- **`frontend/src/lib/api.ts`** — SSE parser in `sendMessage()` already handles `delta`, `tool_call`, and `done` event types. `onDelta` accumulates text, `onToolCall` emits tool name/args.
- **`frontend/src/lib/chat.svelte.ts`** — Reactive state store. `streamingContent` accumulates text deltas. `activeToolCall` tracks current tool. `isStreaming` flag controls display. Tool calls are not collected — only the latest one is stored.
- **`frontend/src/lib/MessageThread.svelte`** — Renders messages. During streaming, shows a temporary assistant bubble with accumulated text. Tool activity shown as a simple label ("Searching PDF…") but no detail or expandability.
- **`internal/api/messages.go`** — Backend sends SSE events: `delta` (text chunk), `tool_call` (name + args), `done`. Tool results are NOT sent to frontend — they're only used in the backend tool loop.
- **`internal/anthropic/client.go`** — Anthropic SSE parsing emits `content_block_delta` and `tool_use` events via channel.
- **DB schema** — Only final `content` text is persisted. Tool use/result details are ephemeral.

### Current gaps

1. Tool call args are sent to frontend but not displayed in detail
2. Tool results are never sent to frontend — backend executes tools and feeds results back to Anthropic silently
3. No expandable/collapsible UI for tool call details
4. No distinction between intermediate streaming text (before tool calls) and final response text (after all tools complete)
5. Multiple tool calls in one response are not tracked — only `activeToolCall` (singular)

## Checklist

- [x] Backend: send `tool_result` SSE events to frontend after tool execution (name, truncated preview, full content)
- [x] Frontend API: parse new `tool_result` SSE event type in `sendMessage()`
- [x] Frontend state: track list of tool interactions (call + result pairs) during streaming, not just single `activeToolCall`
- [x] Frontend UI: render tool call chips inline in the streaming message (tool name + args summary)
- [x] Frontend UI: add expandable popout/modal for tool result content (collapsed by default)
- [x] Frontend UI: show streaming text segments between tool calls (text → tool → text → tool → final text)
- [x] Frontend UI: final message display preserves tool call history as collapsible sections
- [x] Tests: backend test for `tool_result` SSE event emission
- [x] Tests: frontend test for tool interaction state tracking
- [x] Tests: frontend component test for expandable tool result display

## Notes

- Tool results can be large (full page text from `read_page`, search results from `search_pdf`). The popout/modal approach avoids cluttering the message thread.
- `go_to_page` is a client-side tool — its "result" is just a confirmation string, so it can be shown inline without a popout.
- The streaming message should show segments in order: text deltas → tool call chip → (more text if model continues) → next tool call chip → … → final text. This mirrors the actual Anthropic API response structure.
- Consider whether tool call history should be persisted to DB or remain ephemeral. For now, keep it ephemeral (only visible during/after streaming in current session).
