# Task 10: Implement Anthropic Tool Calling for PDF Interactions

Add tool_use support to the Anthropic API client so the assistant can invoke structured actions on the PDF: search for keywords, navigate to pages, read page text into context, and capture page snapshots.

## Context

**Anthropic client** (`internal/anthropic/client.go`):
- Currently sends `apiRequest` with `model`, `max_tokens`, `stream`, `system`, `messages`
- No `tools` field, no tool_use event handling in SSE parser
- Stream parser handles `content_block_delta` (text) and `message_stop` only
- Response channel emits `StreamEvent{Type, Text, Error}`

**Message handler** (`internal/api/messages.go` — `handleSendMessage`):
- Stores user message, loads history, calls `chat.Stream()`, writes SSE deltas, stores final assistant message
- No tool execution loop — streams one response and done

**Frontend streaming** (`frontend/src/lib/api.ts`, `chat.svelte.ts`):
- Reads SSE events `{type:"delta", text}` and `{type:"done"}`
- No handling for tool invocations or frontend-side actions (e.g. page navigation)

**PDF backend** (`internal/pdf/storage.go`):
- File-based storage only (`Save`, `Delete`, `Path`)
- No text extraction or page-level operations on server

**PDF frontend** (`frontend/src/lib/pdf-text.ts`, `pdf-context.svelte.ts`, `PdfViewer.svelte`):
- Browser-side text extraction via pdf.js `extractPageText(page)`
- Page navigation, zoom, intersection-based lazy rendering
- Text selection + surrounding context already wired to chat

**System prompt** (`internal/anthropic/prompt.go`):
- `BuildSystemPrompt(selectedText, surroundingText string)` — only accepts two strings
- Selected text and surrounding context are baked in per-request, not persisted as ongoing state
- No document metadata (title, author, date) or current page number in the prompt
- `Request` struct has `SelectedText` and `SurroundingText` fields but no metadata or page info

**Papers table** (`internal/store/db.go`):
- Only stores `id`, `title`, `file_path`, `file_size`, `created_at`
- No author, publication date, abstract, or other PDF metadata

## Checklist

### Enrich system prompt with document context
- [x] Extract PDF metadata on upload (title, author, subject, creation date, page count) and store in `papers` table (schema migration)
- [x] Extend `anthropic.Request` with `DocumentTitle`, `DocumentAuthor`, `DocumentDate`, `CurrentPage`, `TotalPages` fields
- [x] Rework `BuildSystemPrompt` to accept a context struct instead of loose strings; include document metadata section, current page, selected text, and surrounding text
- [x] Frontend: send `current_page` in the message POST body alongside `selected_text` and `surrounding_text`
- [x] Backend: look up paper metadata from DB in `handleSendMessage` and populate the request context
- [x] Move selected text out of the user message display — it becomes system-prompt-only context, not a visible chat insertion

### Backend: Anthropic client tool_use support
- [x] Extend `apiRequest` with `tools` field and define tool JSON schemas (search_pdf, read_page, go_to_page, snapshot_page)
- [x] Extend `StreamEvent` to carry tool_use blocks (tool name, input, tool_use id)
- [x] Parse `content_block_start` (type=tool_use), `content_block_delta` (type=input_json_delta), and `content_block_stop` SSE events in the stream reader
- [x] Support sending `tool_result` messages back in a multi-turn tool loop

### Backend: PDF text extraction
- [x] Add server-side PDF text extraction (page-level) using a Go PDF library or by calling pdfjs/poppler
- [x] Expose internal helper: `ExtractPageText(pdfPath string, pageNum int) (string, error)`
- [x] Expose internal helper: `SearchText(pdfPath string, query string) ([]SearchResult, error)` returning page numbers and snippets
- [x] Add `PageCount(pdfPath string) (int, error)` helper

### Backend: Tool execution loop in message handler
- [x] After receiving a tool_use event from the stream, pause streaming to the client
- [x] Execute the tool server-side (search_pdf, read_page dispatch to pdf helpers)
- [x] For client-side tools (go_to_page, snapshot_page), send a new SSE event type to the frontend and await a result
- [x] Send `tool_result` back to Anthropic and continue streaming (loop until no more tool_use)
- [x] Store the final assistant text response (excluding intermediate tool calls) in the DB

### Frontend: Handle tool-invocation SSE events
- [x] Add new SSE event types: `{type:"tool_call", name, args}` for client-side tools
- [x] Implement `go_to_page` handler: scroll PDF viewer to requested page
- [x] Implement `snapshot_page` handler: render page to canvas, send image data back to server
- [x] Send tool results back to server (new endpoint or bidirectional channel over the existing SSE stream)

### Frontend: UX for tool activity
- [x] Show indicator in message thread when assistant is executing a tool (e.g. "Searching PDF...")
- [x] Display tool results inline (e.g. search hits as clickable page links)

## Notes

- **Tool schemas** need to follow Anthropic's tool_use format: `{name, description, input_schema}` with JSON Schema for parameters.
- **Multi-turn loop**: The Anthropic API requires sending `tool_result` content blocks in the next user message when the model emits `tool_use`. The server must orchestrate this loop internally before returning the final text to the client.
- **Client-side tools** (go_to_page, snapshot_page) are trickier since the server needs to round-trip with the browser mid-stream. Consider whether snapshot_page could be server-side instead (render PDF page to image using a Go library).
- **Image context**: Anthropic supports `image` content blocks (base64). Snapshot tool results could use this to give the model visual context of charts/figures.
- **Streaming UX**: During the tool loop the user sees no text yet. Need a good loading/activity state.
- **Token budget**: Reading full pages into context can be expensive. Consider truncation or summarization for very long pages.
- **System prompt context**: Selected text, current page, and document metadata should be part of the system prompt so the model always knows what the user is looking at. This replaces the current approach of showing selected text as a visible chat quote — the model just "knows" what's selected.
- **PDF metadata extraction**: Go libraries like `pdfcpu` or `unipdf` can read PDF metadata (Title, Author, Subject, CreationDate from the Info dictionary). Alternatively, the same library used for text extraction can pull this. Extract once at upload time and cache in the DB.
- **Schema migration**: The `papers` table needs new nullable columns (`author`, `subject`, `published_date`, `page_count`) added via `CREATE TABLE IF NOT EXISTS` or an `ALTER TABLE` migration path.
