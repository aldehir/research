# Task 59: Allow pinch gestures to zoom in PDF viewer

Add two-finger pinch-to-zoom support for the PDF viewer on touch devices, using the existing zoom infrastructure.

## Context

Summary of relevant existing code:
- **`src/lib/PdfViewer.svelte`** — Main viewer component with all zoom logic. Ctrl+scroll zoom handled in `handleWheel()` using `zoomByDelta()`. Scale state is component-local (`let scale = $state(1.0)`), clamped to `[0.25, 5.0]`.
- **`src/lib/pdf-utils.ts`** — Zoom utility functions: `zoomByDelta(scale, deltaY)` computes new scale from a delta, `clampScale()`, `fitToWidthScale()`. The `zoomByDelta` function can be reused for pinch gestures.
- **`src/lib/pdf-scroll.ts`** — Scroll anchor preservation during zoom/rerender. Already works with any zoom source.
- **`src/lib/pdf-render.ts`** — Page rendering with scale applied to canvas via pdf.js viewport.
- **No existing touch/pinch handling** — The viewer uses Pointer Events in ResizeHandle and RegionSelect, but no multi-touch tracking exists anywhere.
- The app already sets `touch-action: none` on drag surfaces (ResizeHandle). The PDF scroll container will need `touch-action` managed carefully to allow normal scrolling but intercept two-finger pinch.

## Checklist

- [x] Track pointer events to detect two-finger pinch gestures (pointerdown/pointermove/pointerup with multiple active pointers)
- [x] Compute pinch distance delta and feed into `pinchScale()` to update scale
- [x] Set `touch-action: pan-x pan-y` on the scroll container to prevent browser default pinch-zoom while allowing scroll
- [x] Ensure normal single-finger scroll still works unaffected
- [x] Test pinch-zoom utility functions (pointerDistance, pointerMidpoint, pinchScale)

## Notes

- Uses Pointer Events API (not Touch Events) for consistency with the rest of the codebase.
- Tracks active pointers in a Map keyed by `pointerId`. When exactly 2 pointers are active, computes distance between them; ratio of current/start distance maps to scale factor.
- `touch-action: pan-x pan-y` allows single-finger panning but disables browser pinch-zoom, letting the custom handler take over.
- `pinchScale()` computes `currentScale * (currentDistance / startDistance)`, clamped to [0.25, 5.0].
- The existing `rerenderVisible()` with scroll anchor preservation handles viewport stability during zoom.
