# Task 15: Implement snapshot_page Tool

Add a `snapshot_page` tool that renders a PDF page to a PNG image and sends it to the Anthropic API as an image content block, giving the model visual context for charts, figures, and diagrams.

## Context

The tool infrastructure is already in place from Task 10:
- `internal/anthropic/tools.go` defines tool schemas (`PDFTools()`) — needs a new `snapshot_page` entry
- `internal/api/messages.go` has the tool execution loop (`executeToolCall`) and SSE emission
- `internal/anthropic/client.go` has `ContentBlock` and `Message` types with custom JSON marshaling
- Frontend (`chat.svelte.ts`, `api.ts`, `tool-display.ts`, `MessageThread.svelte`) handles tool call/result display

Key gap: `ContentBlock.Content` is currently a `string`, but Anthropic's API expects image tool results as structured content arrays:
```json
{
  "type": "tool_result",
  "tool_use_id": "toolu_123",
  "content": [
    {
      "type": "image",
      "source": { "type": "base64", "media_type": "image/png", "data": "..." }
    }
  ]
}
```

PDF rendering: No pure-Go PDF rasterizer exists. Use `pdftoppm` (poppler-utils) to shell out and render a page to PNG. System already has poppler libs installed.

## Checklist

- [x] Add `RenderPage(path string, pageNum int) ([]byte, error)` in `internal/pdf/render.go` — shells out to `pdftoppm` to render a single page to PNG
- [x] Test `RenderPage` with a test PDF (verify PNG output, error on invalid page)
- [x] Refactor `ContentBlock.Content` to support both string and structured image content for `tool_result` blocks — update JSON marshaling
- [x] Test that image content blocks serialize correctly for the Anthropic API
- [x] Add `snapshot_page` tool schema to `PDFTools()` in `tools.go`
- [x] Add `snapshot_page` case to `executeToolCall` — call `RenderPage`, base64-encode, return image content block
- [x] Update SSE `tool_result` event to carry content type info (text vs image)
- [x] Update frontend `ToolResult` type and SSE parsing to handle image results
- [x] Update frontend tool result display to render images (e.g. `<img>` tag for image results)
- [x] Update `tool-display.ts` with label/arg formatting for `snapshot_page`

## Notes

- `pdftoppm` is from `poppler-utils` — document as a system dependency. Render at ~150 DPI to balance quality vs token cost (Anthropic charges per image tile).
- The `Content` field refactor is the trickiest part — it changes serialization for all tool_result blocks, so existing tests must keep passing.
- Consider a max page validation to match `read_page` behavior.
- `go_to_page` currently fakes being client-side (returns canned string, SSE triggers navigation as side effect). `snapshot_page` follows the same server-side pattern.
