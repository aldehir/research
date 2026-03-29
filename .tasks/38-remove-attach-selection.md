# Task 38: Remove attach selection from UI and backend

Remove the "attach selected text" feature that lets users quote PDF text selections into chat messages.

## Context

The feature is fully wired end-to-end — contrary to initial assumption, the backend still has it too.

**Frontend files:**
- `frontend/src/lib/MessageInput.svelte` — selection chip UI, quote button, `captureSelection()`/`removeSelection()`, CSS for `.selection-chip`
- `frontend/src/lib/PdfViewer.svelte` — `handleSelectionChange()` listener (lines ~470-486), calls `setSelectedText()`
- `frontend/src/lib/pdf-context.svelte.ts` — `selectedText` state, `setSelectedText()`, `getSelectedText()`, `clearSelectedText()`, `getSurroundingText()`
- `frontend/src/lib/api.ts` — `MessageContext` interface with `selectedText`/`surroundingText`/`currentPage`, included in `sendMessage()` request body
- `frontend/src/lib/chat.svelte.ts` — `sendChatMessage()` accepts optional `context` param

**Backend files:**
- `internal/api/messages.go` — parses `selected_text`/`surrounding_text` from request, `appendViewerContext()` function
- `internal/store/chats.go` — `Message` struct has `SelectedText`/`SurroundingText` fields, stored/retrieved in queries
- `internal/store/db.go` — `messages` table has `selected_text`/`surrounding_text` columns

## Checklist

- [x] Remove selection chip UI and quote button from `MessageInput.svelte`
- [x] Remove `handleSelectionChange` listener and `setSelectedText` call from `PdfViewer.svelte`
- [x] Remove `selectedText`/`clearSelectedText`/`setSelectedText`/`getSelectedText`/`getSurroundingText` from `pdf-context.svelte.ts`
- [x] Remove `MessageContext` interface fields and request body params from `api.ts`
- [x] Remove `context` param from `sendChatMessage()` in `chat.svelte.ts`
- [x] Remove `selected_text`/`surrounding_text` parsing and `appendViewerContext()` from `messages.go`
- [x] Remove `SelectedText`/`SurroundingText` from `Message` struct and queries in `chats.go`
- [x] Drop `selected_text`/`surrounding_text` columns from schema in `db.go` (add migration or just remove from CREATE TABLE if no prod data)
- [x] Verify tests still pass (`go test ./...` and `pnpm test`)

## Notes

- The `current_page` field in the API request may still be useful for viewer context without selected text — decide whether to keep it.
- Database columns can simply be removed from the CREATE TABLE since this is a dev tool with no production migration concerns.
