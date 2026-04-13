# Task 62: Paste and drop images into chat

Allow users to paste images from clipboard or drag-and-drop image files into the chat input, sending them to the LLM as image content blocks.

## Context

Image attachment infrastructure already exists from tasks 49/51:
- `MessageAttachment` type (`api.ts:175`): `{ image_data, text, page }`
- `addAttachment` / `consumeAttachments` in `attachments.svelte.ts`
- Attachment thumbnail strip + preview modal in `MessageInput.svelte`
- Backend `buildMultimodalUserMessage` (`messages.go:419`) already handles empty `text` and appends `PartImage` blocks
- Anthropic adapter converts `chat.PartImage` to `image` content blocks (`convert.go`)
- Attachment persistence to disk + DB already works

Key things to change:
- **`MessageInput.svelte`**: Add `paste` handler on textarea, `dragover`/`drop` on input area, resize images client-side
- **`MessageInput.svelte`**: Allow sending when attachments exist but text is empty (`handleSend` line 28 blocks on `!content`)
- **Attachment strip**: Handle `page=0` label (pasted images have no source page)
- **No mobile additions** — no camera/gallery button

## Plan

All work is frontend-only. The backend already supports the message shape.

1. Add a `resizeImage(file: File | Blob): Promise<string>` utility that draws to a canvas capped at 2048px on longest side, returns base64 PNG (no `data:` prefix).
2. Add `paste` event handler on the textarea — extract image from `clipboardData.items`, resize, call `addAttachment({ image_data, text: '', page: 0 })`.
3. Add `dragover`/`dragleave`/`drop` handlers on `.input-area` — accept image files, resize, add as attachments. Show a visual drop indicator.
4. Fix `handleSend` to allow sending when `attachments.length > 0` even if text is empty.
5. Fix attachment strip label: show "Pasted image" or similar instead of "p.0" when `page === 0`.
6. Tests for the resize utility and send-guard logic.

## Checklist

- [x] `resizeImage` utility: canvas-based resize to max 2048px, returns base64 PNG — unit test with mock canvas
- [x] Paste handler on textarea: extract image from clipboard, resize, add attachment — test paste event flow
- [x] Drag-and-drop on input area: dragover/dragleave/drop handlers with visual indicator — test drop event flow
- [x] Fix `handleSend` to allow image-only messages (no text required)
- [x] Fix attachment thumbnail label for `page === 0` (no source page)
- [x] Integration test: pasted image appears in attachment strip and sends correctly

## Notes

- Resize to max 2048px on longest side to keep payloads reasonable (Anthropic supports up to ~20MB but large images are slow)
- Convert everything to PNG on the canvas
- No mobile-specific UI (no camera/gallery button)
- `page: 0` and `text: ''` are valid — backend already skips the `[Attached region from page X]` annotation when text is empty
