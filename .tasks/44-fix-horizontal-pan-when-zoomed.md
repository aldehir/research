# Task 44: Fix horizontal panning when zoomed in

When the PDF is zoomed in so pages are wider than the viewport, the user can
scroll right to see content beyond the right edge but cannot scroll left —
content past the left edge is hidden and unreachable.

## Context

**Root cause**: `.pages-container` in `PdfViewer.svelte` uses `align-items: center`.
When flex children (pages) are narrower than the container this centers them
nicely. But when they are *wider*, flex places the overflow equally on both
sides — left and right. Since the container's horizontal scroll starts at 0,
the left overflow is behind the scroll origin and can never be reached.

**Fix**: Replace `align-items: center` on `.pages-container` with
`align-items: flex-start`, and add `margin-inline: auto` on `.page-wrapper`.
Auto margins center the page when it fits within the container, but collapse
to 0 when the page overflows — so the left edge is always anchored at scroll
position 0 and the full width is scrollable.

This is more robust than `align-items: safe center`, which has had
inconsistent behavior in some Safari/Chrome versions.

**Key files**:
- `frontend/src/lib/PdfViewer.svelte` — `.pages-container` and `.page-wrapper`
  CSS rules

## Checklist

- [x] Change `align-items: center` to `align-items: flex-start` on `.pages-container`
- [x] Add `margin-inline: auto` to `.page-wrapper`

## Notes

- No JS changes needed — this is a pure CSS fix.
- The mobile toolbar wrapping behaviour is unrelated and should not be affected.
