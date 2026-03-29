# Task 43: Improve Resize Handles for Touch Devices

Make panel resize handles usable on iPad and other touchscreen devices by expanding the
hit target and adding a visible affordance for coarse-pointer inputs.

## Context

`frontend/src/lib/ResizeHandle.svelte` — the single reusable handle used by:
- `+layout.svelte` (sidebar left, chat right)
- `PdfViewer.svelte` (ToC panel)

The handle uses Pointer Events API (already touch-compatible in principle), but the
hit target is only **6px wide** with a **2px visible indicator** — far below the 44px
minimum recommended for touch. On coarse-pointer devices hover states never fire, so
there is also no visual affordance that the handle exists.

Existing tests live in:
- `frontend/tests/resize-handle.test.ts` — delta logic
- `frontend/tests/panel-resize.test.ts` — width constraints
- `frontend/tests/panel-widths.test.ts` — reactive state

The fix is CSS-only (plus a grabber element in the template) — no changes to pointer
event logic or state management are needed.

## Checklist

- [x] Expand touch hit target on coarse-pointer devices (`@media (pointer: coarse)`) to
      at least 44px via negative horizontal margins or a wider handle div
- [x] Show the resize indicator in its resting (visible) state on coarse-pointer devices
      so the handle is discoverable without hover
- [x] Add a grabber affordance (e.g. short vertical dots/lines centered on the handle)
      that is visible on coarse-pointer devices and hidden on fine-pointer devices
- [x] Verify no regression in existing visual layout on desktop (fine pointer)

## Notes

- Use `@media (pointer: coarse)` rather than `(hover: none)` — iPads report no-hover
  but fine/coarse pointer is the more reliable signal for hit-target sizing.
- The negative-margin trick keeps the visible indicator narrow while making the
  transparent clickable zone wide: `margin-inline: -19px; padding-inline: 19px` on a
  44px-wide handle keeps layout unchanged.
- No new dependencies — CSS and inline SVG/pseudo-elements only.
- No backend or state-management changes required.
