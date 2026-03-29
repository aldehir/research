# Task 50: Defer chat persistence until first message

Don't persist new chat sessions to the database until the user sends their first message. Currently clicking the "+" button immediately creates a row in `chat_sessions`, which accumulates empty chats.

## Context

Current flow when "+" is clicked:
1. `ChatPanel.svelte:28-31` — `handleNew()` calls `createSession(paperId)`
2. `chat.svelte.ts:34-39` — `createSession()` calls `createChatSession()` API client
3. `api.ts:88-99` — `POST /api/papers/{paperId}/chats`
4. `internal/api/chats.go:40-88` — handler generates UUID, title, inserts via `store.CreateChatSession()`
5. `internal/store/chats.go:29-35` — SQL INSERT into `chat_sessions`

Message sending (`chat.svelte.ts:74-152`, `api.ts:170-253`) assumes the chat already exists in the DB — it posts to `POST /api/papers/{paperId}/chats/{chatId}/messages` and the backend loads history from the existing session.

Key files:
- `frontend/src/lib/chat.svelte.ts` — session state, `createSession`, `sendChatMessage`
- `frontend/src/lib/ChatPanel.svelte` — "+" button, session dropdown
- `frontend/src/lib/api.ts` — `createChatSession`, `sendMessage`
- `internal/api/chats.go` — `handleCreateChatSession`
- `internal/api/messages.go` — `handleSendMessage`
- `internal/store/chats.go` — `CreateChatSession`, `ListChatSessions`

## Approach

Introduce a "draft" session concept on the frontend. When "+" is clicked, create a local-only session object (with a temporary ID). On the first `sendChatMessage`, call the create-session API first, swap the temp ID for the real one, then send the message. The backend stays unchanged.

## Checklist

- [x] Frontend: `createSession` creates a local draft session (no API call)
- [x] Frontend: `sendChatMessage` detects draft session, calls create API before sending message
- [x] Frontend: session dropdown and active state work correctly with draft sessions
- [x] Frontend: switching away from an unsent draft discards it
- [x] Test: unit test that createSession does not call fetch
- [x] Test: unit test that first message in draft triggers session creation then message send
- [x] Test: switching sessions discards empty draft

## Notes

- The backend `POST /api/papers/{paperId}/chats` and `POST .../messages` endpoints stay as-is.
- Draft sessions need a distinguishable temp ID (e.g. `draft-` prefix) to avoid collision with real UUIDs.
- When listing sessions from API on paper load, drafts won't appear (they're local-only), which is the desired behavior.
- If user clicks "+" again while a draft exists, reuse the existing draft rather than creating another.
