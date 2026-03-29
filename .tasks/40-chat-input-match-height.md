# Task 40: Match textarea and send button height in chat input

Make the chat input textarea and send button the same height so they appear visually aligned.

## Context

- **MessageInput component**: `frontend/src/lib/MessageInput.svelte`
- The `.input-row` uses `display: flex` with `align-items: flex-end`
- **Textarea** height is implicitly set by `rows="2"` — no explicit CSS height
- **Send button** has `height: var(--btn-height-lg)` (36px), much shorter than the 2-row textarea
- On mobile (≤1023px) the button gets `min-width: 44px; min-height: 44px`
- Theme tokens in `frontend/src/lib/theme.css`

## Checklist

- [x] Set send button height to match the textarea (remove fixed `--btn-height-lg`, use `align-items: stretch` or explicit matching height)
- [x] Verify alignment looks correct in both light and dark themes
- [x] Verify mobile layout still meets touch-target minimums

## Notes

- Simplest approach: change `.input-row` to `align-items: stretch` so the button stretches to match the textarea height, and remove the explicit height on `.send-btn`.
- Alternative: set both to an explicit shared height and drop `rows="2"`.
