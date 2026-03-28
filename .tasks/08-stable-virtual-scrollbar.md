# Task 08: Stabilize scrollbar for virtual PDF rendering

The PDF viewer uses virtual rendering (IntersectionObserver) to only render visible pages. The native scrollbar can resize or jump because:
1. Placeholder dimensions may not exactly match rendered page dimensions (rounding, canvas scaling)
2. During zoom/rerender, all pages are cleared and resized simultaneously, causing scroll position loss
3. No scroll anchoring — when content above the viewport changes height, the scroll position shifts

## Context

Key files:
- `frontend/src/lib/PdfViewer.svelte` — main viewer with IntersectionObserver-based virtual rendering, scroll handling, zoom
- `frontend/src/lib/pdf-render.ts` — `renderPage()`, `clearPage()`, `getPageDimensions()`, `PDF_TO_CSS_UNITS`
- `frontend/src/lib/pdf-utils.ts` — zoom/scale utilities

Current approach:
- All page DOM elements are created upfront as empty `<div class="page-wrapper">` placeholders
- `getPageDimensions()` sets placeholder width/height based on scale
- `renderPage()` creates canvas at the computed viewport size — if this differs from placeholder, scroll jumps
- `clearPage()` removes rendered content; `handleIntersection` restores placeholder dims
- `rerenderVisible()` clears all pages and resizes placeholders, then re-triggers observer — no scroll position preservation
- Scroll container uses native `overflow: auto` with no anchoring

## Checklist

- [x] Ensure placeholder dimensions exactly match rendered canvas dimensions (audit `getPageDimensions` vs `renderPage` viewport calculation)
- [x] Preserve scroll position during zoom/rerender by anchoring to the page currently in view
- [x] Prevent scrollbar resize flicker when pages enter/leave the rendered set (dimensions must be stable across render/clear cycles)
- [x] Test: placeholder dims match rendered dims for various scales
- [x] Test: scroll position is preserved (relative to current page) after zoom in/out
- [x] Test: scrollbar thumb size remains constant while scrolling through a document

## Notes

- CSS `overflow-anchor` could help prevent jump during content changes, but browser support and interaction with IntersectionObserver needs investigation
- The scroll anchor during zoom should record which page is at viewport center (or top) and its offset, then restore after rerender
- Consider whether `rerenderVisible()` should batch dimension updates vs the current clear-all-then-re-observe approach
- Must not break fit-to-width mode or keyboard/wheel navigation
