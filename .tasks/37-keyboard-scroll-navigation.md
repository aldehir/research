# Task 37: Keyboard scroll navigation in PDF viewer

Add keyboard shortcuts for smooth scroll-based navigation: Space for half-page down, arrow keys for small jumps, and Page Up/Down for half-page jumps (replacing current full-page snapping behavior).

## Context

- `frontend/src/lib/PdfViewer.svelte` has the `handleKeydown()` function (line ~409) that handles keyboard events
- Current PageUp/PageDown behavior calls `goToPage()` which does `scrollIntoView()` to snap to page boundaries — needs to change to half-viewport scrolling instead
- The scroll container is `.pages-container` bound to `scrollContainer` variable
- Arrow keys and Space are not currently handled
- Handler already skips events when focus is in an INPUT element (line 411) — also need to skip for TEXTAREA
- No existing tests for keyboard navigation; `pdf-utils.ts` has unit tests for zoom/clamp helpers

## Checklist

- [x] Test and implement Space key: scroll down by half the container height (smooth scroll)
- [x] Test and implement ArrowDown/ArrowUp: scroll by a small increment (~100px) down/up
- [x] Test and implement PageDown/PageUp: scroll by half the container height (replace current goToPage behavior)
- [x] Ensure Space, arrows, and PageDown/PageUp all call `e.preventDefault()` to avoid browser defaults
- [x] Skip handling when focus is in INPUT or TEXTAREA elements

## Notes

- Use `scrollContainer.scrollBy({ top: delta, behavior: 'smooth' })` for relative scrolling rather than `scrollIntoView`
- Half-page = `scrollContainer.clientHeight / 2`
- Small jump = ~100px (tune if needed)
- Home/End behavior (jump to first/last page) stays as-is
