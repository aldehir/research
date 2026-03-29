# Task 34: Resizable ToC panel

Make the Table of Contents panel within PdfViewer resizable using the same drag-handle pattern as the sidebar and chat panels.

## Context

Summary of relevant existing code:

- **TocPanel.svelte** — Renders the ToC tree. Currently has no width prop; its width is set by the parent `PdfViewer.svelte` via `.viewer-body :global(.toc-panel)` at a fixed `260px`.
- **PdfViewer.svelte** — Hosts the ToC inside `.viewer-body` (a flex row): `[TocPanel (260px fixed)] [.pages-container (flex:1)]`. ToC visibility toggled by `tocVisible` state.
- **ResizeHandle.svelte** — Reusable drag handle (pointer-capture based, 6px wide, shows primary-color indicator). Accepts `onResize(delta)` and `side` prop. Already used for sidebar and chat.
- **panel-resize.ts** — Constraints and persistence. Defines min/max/default constants, `clampWidth`, `clampResize`, `savePanelWidths`/`loadPanelWidths`. Currently only handles `sidebar` and `chat` in `PanelWidths`.
- **panel-widths.svelte.ts** — Reactive state module with getters/setters/handlers for sidebar and chat widths. Calls `clampResize` and `savePanelWidths` on resize.
- **panel-resize.test.ts** / **panel-widths.test.ts** — Test suites for the constraint logic and reactive state.

Key patterns to follow:
- Add `TOC_MIN`, `TOC_MAX`, `TOC_DEFAULT` constants to `panel-resize.ts`
- Extend `PanelWidths` to include `toc` and update save/load
- Add `tocWidth` state + getter/setter/handler to `panel-widths.svelte.ts`
- Place a `<ResizeHandle>` between TocPanel and `.pages-container` in PdfViewer
- ToC resize is scoped within the center content area (doesn't interact with sidebar/chat constraints directly)

## Checklist

- [x] Add TOC constants and extend `clampResize` / `PanelWidths` in `panel-resize.ts` + tests
- [x] Add `tocWidth` state, getter/setter, `handleTocResize`, and init/save to `panel-widths.svelte.ts` + tests
- [x] Place `ResizeHandle` between TocPanel and pages-container in `PdfViewer.svelte`, wire up resize handler
- [x] Pass width to TocPanel (via style or prop) instead of fixed CSS `260px`
- [x] Verify persistence: ToC width saves/restores from localStorage
- [x] Verify mobile: no resize handle shown on small screens (< 1024px)

## Notes

- The ToC panel lives inside the center content area (between sidebar and chat), so its max width should be constrained by the available center width, not the full viewport.
- The `clampResize` function currently only knows about `sidebar` | `chat`. We can either extend it or add a simpler clamping path for the ToC since it doesn't compete with sidebar/chat directly — it just needs min/max and a reasonable upper bound from the center area width.
- The ResizeHandle `side` prop should be `'left'` (default) since the handle is to the right of the ToC, and dragging right should grow the panel.
