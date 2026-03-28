# Task 07: Attach selected text and page content to chat messages

Wire up PDF text selection to the chat input so selected text is sent with messages and displayed as context. Also send the text content of the previous, current, and next pages as surrounding context in the system prompt.

## Context

Summary of relevant existing code:

- **Backend already supports `selected_text` and `surrounding_text`** fields end-to-end:
  - `internal/api/messages.go:35-38` — request body parses both fields
  - `internal/store/chats.go` — `messages` table has nullable `selected_text` and `surrounding_text` columns
  - `internal/anthropic/prompt.go` — `BuildSystemPrompt()` appends selected/surrounding text to base prompt
  - `internal/anthropic/client.go` — `Request` struct carries `SelectedText` and `SurroundingText`

- **Frontend gaps:**
  - `api.ts:103-110` — `sendMessage()` only sends `content`, ignores `selected_text`/`surrounding_text`
  - `api.ts:59-64` — `Message` interface lacks `selected_text` field
  - `chat.svelte.ts:47-50` — `sendChatMessage()` only takes `content` string
  - `MessageInput.svelte:25` — calls `sendChatMessage` with content only
  - `MessageThread.svelte:28-30` — already renders `selected_text` as blockquote (ready to use)

- **PDF text extraction**: pdf.js `page.getTextContent()` is already called in `pdf-render.ts:90` for TextLayer rendering. Can reuse to extract plain text from any page.

- **Page tracking**: `PdfViewer.svelte` tracks `currentPage` (1-indexed) and holds all `PDFPageProxy` objects in `pages` array.

## Checklist

- [x] Add `selected_text` to frontend `Message` interface in `api.ts`
- [x] Update `sendMessage()` in `api.ts` to accept and send `selected_text` and `surrounding_text`
- [x] Add page text extraction utility (extract plain text from a `PDFPageProxy` via `getTextContent()`)
- [x] Expose PDF text selection state from `PdfViewer.svelte` (capture `window.getSelection()` within PDF container)
- [x] Expose a function to get text content of pages by number from `PdfViewer.svelte`
- [x] Update `sendChatMessage()` in `chat.svelte.ts` to accept `selected_text` and `surrounding_text`
- [x] Wire `MessageInput.svelte` to read selected text from PDF and extract prev/current/next page text on send
- [x] Show selected text chip/preview in input area before sending (UX feedback)
- [x] Verify `MessageThread.svelte` blockquote rendering works end-to-end

## Notes

- Page text extraction should happen client-side using pdf.js (already loaded) rather than adding a Go PDF parsing dependency.
- `surrounding_text` will contain concatenated text of pages [current-1, current, current+1], clamped to valid range.
- The selected text capture should use the native browser selection API within the PDF container's `.textLayer` elements.
- Consider clearing the selection indicator after sending a message.
- The `selected_text` on user messages is for display; the `surrounding_text` (page content) is for the system prompt only and doesn't need to be stored/displayed.
