# Task 57: Redesign PDF toolbar for mobile

The PDF toolbar (`PdfViewer.svelte`) is too wide on mobile and wraps across multiple rows due to `flex-wrap: wrap` combined with 44px touch targets. It needs to be reimagined to fit comfortably on small screens.

## Context

The toolbar lives in `frontend/src/lib/PdfViewer.svelte` (lines 549–597) and contains two groups:

**Left group (navigation):**
- ToC toggle button
- Prev/Next page buttons
- Page info text (`"5 / 120"`)
- "Go to page" input

**Right group (zoom & tools):**
- Zoom out/in buttons + zoom display
- Fit to width button
- Region selection button

**Current mobile CSS** (lines 780–795): `flex-wrap: wrap` with `min-width/height: 44px` buttons causes multi-row wrapping on narrow viewports.

**Desktop CSS** (lines 646–655): `display: flex; justify-content: space-between` with `0.75rem` gap.

### Related files
- `frontend/src/lib/PdfViewer.svelte` — toolbar markup and styles
- `frontend/src/lib/mobile-layout.svelte.ts` — mobile state management
- `frontend/src/routes/+layout.svelte` — app header (above toolbar)
- `frontend/src/lib/theme.css` — CSS custom properties

## Checklist

- [x] Design a compact mobile toolbar layout (e.g. collapse into icon menus, overflow popover, or split into contextual rows)
- [x] Test: toolbar renders without wrapping on 375px-wide viewport
- [x] Test: all toolbar actions remain accessible on mobile
- [x] Test: touch targets remain at least 44px
- [x] Implement the new mobile toolbar layout
- [x] Verify desktop toolbar is unaffected

## Notes

- Possible approaches: overflow menu/popover for less-used controls, swipeable toolbar segments, or a condensed single-row with only essential controls visible
- The "Go to page" input and zoom controls are lower-frequency actions — good candidates for an overflow menu
- Must keep ToC toggle, page nav, and region select easily accessible
- Consider whether the page info text can be made more compact (e.g. just "5/120")
