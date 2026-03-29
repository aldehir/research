# Task 51: Persist chat images to disk and recall on conversation load

Store images attached to chat messages (from region selections) on disk and reference them in SQLite so they survive page reloads and appear when reopening a conversation.

## Context

Current flow: user selects a PDF region → `POST /api/papers/{id}/region` returns `{image_data, text}` as base64 PNG → frontend holds it in a reactive `messageAttachments` Map → sent inline in the `POST .../messages` JSON body → backend builds multimodal Anthropic content blocks in memory → image is never persisted.

When a chat is reloaded via `GET /api/papers/{id}/chats/{chatId}`, messages come back with text only — `getUserAttachments(messageId)` returns `undefined` because the Map is empty.

### Key files

**Frontend:**
- `frontend/src/lib/attachments.svelte.ts` — pending attachment state (`PendingAttachment` with `image_data`, `text`, `page`)
- `frontend/src/lib/chat.svelte.ts:74-151` — `sendChatMessage()` stores attachments in `messageAttachments` Map keyed by message ID
- `frontend/src/lib/api.ts:164-168` — `MessageAttachment` interface (`image_data`, `text`, `page`)
- `frontend/src/lib/MessageThread.svelte:110-133` — renders user attachments via `getUserAttachments()`

**Backend:**
- `internal/api/messages.go:40-58` — parses `attachment{ImageData, Text, Page}` from request body
- `internal/api/messages.go:97-146` — builds multimodal Anthropic message blocks (text + image content blocks)
- `internal/store/chats.go:14-20` — `Message` struct (no attachment fields)
- `internal/store/db.go:65-71` — `messages` table schema (no attachment columns)
- `internal/pdf/storage.go` — existing `Storage` struct pattern: `{Dir}/{id}.pdf`

### What's missing
- No `message_attachments` table in SQLite
- No disk storage for attachment images
- No API to retrieve attachment images
- Frontend doesn't request attachments when loading chat history

## Checklist

- [x] DB: add `message_attachments` table (id, message_id, file_path, text, page, created_at) with migration
- [x] Store: add `CreateAttachment`, `ListAttachmentsByMessage`, `ListAttachmentsByChat` functions
- [x] Backend: save attachment images to `{data_dir}/attachments/{id}.png` when message is sent
- [x] Backend: persist attachment metadata to `message_attachments` table in `handleSendMessage`
- [x] API: add `GET /api/attachments/{id}/image` endpoint to serve stored PNGs
- [x] API: include attachment metadata when returning messages in `GET /api/papers/{id}/chats/{chatId}`
- [x] Frontend: update `Message` type to include optional attachments array from API
- [x] Frontend: render persisted attachments on chat history load (replace ephemeral Map lookup)
- [x] Frontend: continue using base64 for optimistic display before persistence round-trips
- [x] Test: store function tests for CreateAttachment and ListAttachmentsByMessage
- [x] Test: handler test for attachment persistence during message send
- [x] Test: handler test for attachment image serving endpoint

## Notes

- Images are base64-encoded PNGs from `pdf.RenderRegion()`. Decode to raw PNG bytes before writing to disk.
- Serve stored images via a dedicated endpoint rather than inlining base64 in JSON responses — keeps chat history payloads small.
- The `message_attachments` table references `messages(id)` with `ON DELETE CASCADE` (follows existing pattern).
- Anthropic API still receives base64 inline (no change to LLM flow) — persistence is orthogonal.
- Tool-generated images (e.g. `snapshot_page` results) are handled by task 53 (persist tool interactions). The image storage mechanism from this task should be reusable for tool result images.
- The frontend `messageAttachments` Map can be kept for optimistic rendering of just-sent messages, with persisted URLs taking over on next load.
