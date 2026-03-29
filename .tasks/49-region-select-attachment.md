# Task 49: Region selection to attach PDF excerpt to chat

Select a rectangular region on the PDF, extract both an image (pdftoppm) and text (pdftotext) from that region, and auto-attach it to the chat message input. The user can preview, expand, or dismiss the attachment before sending.

## Context

Summary of relevant existing code:

- **PDF rendering** (`frontend/src/lib/pdf-render.ts`): Pages render to canvas at `scale * PDF_TO_CSS_UNITS` (96/72). Each page lives in a `.page-wrapper` div managed by `PdfViewer.svelte`. Coordinate conversion from screen pixels to PDF points: `points = pixels / (PDF_TO_CSS_UNITS * scale)`.
- **PDF viewer** (`frontend/src/lib/PdfViewer.svelte`): Toolbar with zoom/nav buttons, `.pages-container` with scroll, lazy rendering via IntersectionObserver. Pages tracked in `pageElements` Map. Already has keyboard/mouse handlers for zoom.
- **Server-side extraction** (`internal/pdf/text.go`): `ExtractPageText()` calls `pdftotext -layout -f N -l N`. Both `pdftotext` and `pdftoppm` support `-x -y -W -H` flags for bounding region (coordinates in PDF points).
- **Server-side rendering** (`internal/pdf/render.go`): `RenderPage()` calls `pdftoppm -png -r 150 -f N -l N -singlefile`, then crops whitespace and constrains to 1568px max.
- **Anthropic client** (`internal/anthropic/client.go`): `Message` supports `ContentBlocks` with `ContentPart` arrays including `type: "image"` with base64 source. Already used by `snapshot_page` tool results.
- **Chat message flow** (`internal/api/messages.go`): `handleSendMessage` parses `{ content, current_page }`, stores user message, builds Anthropic messages, streams response. Tool results already handle image content parts.
- **Frontend message input** (`frontend/src/lib/MessageInput.svelte`): Simple textarea + send button. No attachment support yet.
- **Chat store** (`frontend/src/lib/chat.svelte.ts`): `sendChatMessage()` passes content + currentPage to `sendMessage()` API function.
- **API client** (`frontend/src/lib/api.ts`): `sendMessage()` posts `{ content, current_page }` and reads SSE stream.
- **Icons** (`frontend/src/lib/icons/index.ts`): Lucide-sourced SVG path constants. Need to add a selection/crosshair icon (e.g. `BoxSelect` or `Crosshair` from Lucide).

## Checklist

### Backend: region extraction

- [x] Add `ExtractRegionText(path, page, x, y, w, h)` to `internal/pdf/text.go` — calls `pdftotext -layout -f page -l page -x X -y Y -W W -H H`
- [x] Add `RenderRegion(path, page, x, y, w, h)` to `internal/pdf/render.go` — calls `pdftoppm -png -r 150 -f page -l page -x X -y Y -W W -H H -singlefile`, constrains size
- [x] Add `POST /api/papers/{id}/region` handler — accepts `{ page, x, y, w, h }` (PDF points), returns `{ text, image_data }` (base64 PNG)
- [x] Register route in `internal/api/api.go`

### Backend: multimodal user messages

- [x] Extend `handleSendMessage` to accept optional `attachments` array in request body: `[{ image_data, text, page }]`
- [x] When attachments present, build Anthropic user message with content blocks: text block (user message + extracted text) + image block (base64 PNG)

### Frontend: region selection overlay

- [x] Add selection icon to `frontend/src/lib/icons/index.ts` (Lucide `BoxSelect` or similar)
- [x] Create `frontend/src/lib/RegionSelect.svelte` — overlay on pages-container that captures mousedown/mousemove/mouseup to draw a selection rectangle; converts pixel coords to PDF points using page element offset + scale; emits `{ page, x, y, w, h }` on completion
- [x] Add toolbar toggle button in `PdfViewer.svelte` for selection mode; when active, render `RegionSelect` overlay; disable text selection and scroll-zoom while selecting
- [x] Handle touch events in RegionSelect for tablet support

### Frontend: attachment in message input

- [x] Add `extractRegion(paperId, page, x, y, w, h)` to `frontend/src/lib/api.ts`
- [x] Create attachment state (image_data + text + page) — either in `MessageInput.svelte` local state or a shared store callable from PdfViewer
- [x] When region selected, call `extractRegion`, store result as pending attachment
- [x] Show attachment thumbnail strip above textarea in `MessageInput.svelte` — small image preview, page number label, expand button (shows full image + text in a popover/modal), dismiss (X) button
- [x] Thread attachment through `sendChatMessage` in `chat.svelte.ts` and `sendMessage` in `api.ts` — include `attachments` in POST body
- [x] Display attachment image inline in user message bubble in `MessageThread.svelte`

## Notes

- **Coordinate system**: `pdftotext` and `pdftoppm` both use PDF points (1/72 inch) for `-x -y -W -H`. The page origin is top-left. Frontend conversion: `pdf_points = screen_pixels / (PDF_TO_CSS_UNITS * scale)` where `PDF_TO_CSS_UNITS = 96/72`.
- **Auto-attach**: After drawing the rectangle, immediately call the extraction endpoint and attach the result — no confirmation dialog. User can dismiss via the X on the thumbnail.
- **Ephemeral image**: The base64 image data is sent to Anthropic but not persisted in the messages DB. On reload, user messages with attachments will show the text but not the image (same pattern as `snapshot_page` tool results today).
- **Selection mode exit**: After a successful region selection, automatically exit selection mode so the user returns to normal PDF interaction.
- **DPI**: Use 150 DPI for pdftoppm region rendering (matches existing `RenderPage`). Skip whitespace cropping for regions since the user explicitly chose the bounds.
- **Error handling**: If the region has no extractable text, still attach the image — the visual content is the primary value. Show empty text gracefully.
