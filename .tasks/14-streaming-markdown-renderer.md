# Task 14: Add streaming markdown renderer to chat panel

Render assistant messages as formatted markdown instead of plain text, supporting incremental rendering during SSE streaming.

## Context

Summary of relevant existing code:

- **MessageThread.svelte** renders messages. Assistant text currently displayed via `{segment.content}` or `{message.content}` with `white-space: pre-wrap` — no formatting at all.
- **chat.svelte.ts** manages streaming state. `streamSegments` is an array of `StreamSegment` (text or tool), updated incrementally via `appendTextSegment()` as SSE deltas arrive.
- **api.ts** `sendMessage()` handles SSE parsing — each `delta` event calls `onDelta(text)` which appends to `streamingContent` and the current text segment.
- No markdown library is installed. `pdfjs-dist` is the only runtime dependency.
- Only assistant messages need markdown rendering; user messages stay plain text.
- Streaming deltas arrive as small text chunks — the renderer must handle incomplete markdown gracefully (e.g., partial code fences, half-finished lists).

## Checklist

- [ ] Choose and install a markdown parsing library (e.g., `marked`, `markdown-it`) — evaluate streaming-friendly options
- [ ] Create `MarkdownRenderer.svelte` component that takes a `content: string` prop and renders sanitized HTML
- [ ] Write tests for basic markdown rendering (headings, bold, italic, code spans, links)
- [ ] Write tests for code block rendering with syntax highlighting consideration
- [ ] Write tests for list rendering (ordered, unordered, nested)
- [ ] Handle streaming/incremental content — ensure partial markdown doesn't break layout (e.g., unclosed code fences during typing)
- [ ] Integrate into MessageThread.svelte for assistant message segments (both completed and streaming)
- [ ] Add CSS styling for markdown elements (code blocks, blockquotes, tables, lists) that fits the chat panel layout
- [ ] Sanitize HTML output to prevent XSS from model-generated content
- [ ] Verify user messages remain plain text (no markdown rendering)
- [ ] Test with real streaming responses end-to-end

## Notes

- The renderer must be resilient to incomplete markdown since streaming chunks can split mid-syntax (e.g., receiving `` ``` `` without the closing fence yet). Re-rendering the full accumulated text on each delta is the simplest approach.
- Code blocks from LLM responses are common — syntax highlighting (e.g., via `highlight.js` or `shiki`) would be a nice addition but could be deferred.
- XSS sanitization is critical since we're rendering model output as HTML. Use a library like `DOMPurify` or the markdown library's built-in sanitization.
- Consider performance: re-parsing the full message on every delta could be expensive for long responses. Profile and optimize if needed.
- Markdown elements need styling that works within the existing chat bubble layout (max-width: 90%, 0.9rem font).
