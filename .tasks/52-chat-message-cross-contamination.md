# Task 52: Fix chat message cross-contamination when switching sessions

When sending a message in one chat and switching to another chat while streaming, the second chat displays streaming content from the first chat.

## Context

The root cause is global (module-level) streaming state in `frontend/src/lib/chat.svelte.ts`:

- `messages`, `streamingContent`, `streamSegments`, `isStreaming` are singular module-level `$state` variables shared across all chat sessions.
- `sendChatMessage()` starts SSE streaming, and `onDelta`/`onToolCall` callbacks write to these globals.
- `selectSession()` overwrites `messages` with the new chat's messages but does NOT cancel the active stream or clear streaming state.
- `MessageThread.svelte` renders from these same globals via `getMessages()`, `getIsStreaming()`, `getStreamSegments()` — no chat-specific filtering.

**Sequence:**
1. User sends message in Chat A → streaming starts, callbacks write to global `streamingContent`
2. User switches to Chat B → `messages` overwritten with Chat B's messages
3. Chat A's `onDelta` callbacks continue firing → `streamingContent` accumulates Chat A content
4. Chat B now renders Chat A's streaming text

**Backend is NOT affected** — messages are correctly keyed by `chat_session_id` in SQLite.

Key files:
- `frontend/src/lib/chat.svelte.ts` — global streaming state (lines 16-24), `sendChatMessage()` (lines 92-179), `selectSession()` (lines 56-61)
- `frontend/src/lib/MessageThread.svelte` — renders globals without chat filtering
- `frontend/src/lib/ChatPanel.svelte` — triggers session switching
- `frontend/src/lib/api.ts` — `sendMessage()` SSE client, returns `AbortController`

## Checklist

- [x] Test: switching chat during streaming does not show stale streaming content
- [x] Store active stream's chat ID and abort/clear streaming state on session switch
- [x] Test: completed message from previous stream is saved to the correct chat session
- [x] Ensure `onDone` callback no-ops or routes to correct session if chat was switched mid-stream

## Notes

- `sendMessage()` in `api.ts` returns an `AbortController` — can use this to cancel the SSE connection on chat switch.
- Simplest fix: track which `chatId` owns the active stream; in `selectSession()`, abort the stream and clear streaming state. The `onDone` callback should check if the stream's chat still matches `activeSessionId` before appending the assistant message.
- Alternative: let the stream finish in the background but suppress rendering — more complex, less predictable UX.
