# Task 23: Add resize handles to side panels in desktop view

Add draggable resize handles between the side panels and center content so users can adjust panel widths in desktop view (≥ 1024px).

## Context

The layout is a three-panel flexbox row in `frontend/src/routes/+page.svelte`:
- **Left sidebar** (`.sidebar`): fixed 280px, min-width 220px — contains paper list
- **Center content** (`.content`): flex: 1 — contains PDF viewer
- **Right chat panel** (`.chat-panel` in `ChatPanel.svelte`): fixed 360px, min-width 300px

Additionally, the PDF viewer has an internal **TOC panel** (`.toc-panel` in `TocPanel.svelte`): 260px, min-width 200px.

No resize or drag-handle patterns exist in the codebase. Panel widths are currently hardcoded in CSS. The mobile layout (< 1024px) uses slide-over overlays and should not be affected.

Key files:
- `frontend/src/routes/+page.svelte` — main layout container
- `frontend/src/lib/ChatPanel.svelte` — right panel (360px)
- `frontend/src/lib/PdfViewer.svelte` — center content with internal TOC panel
- `frontend/src/lib/TocPanel.svelte` — TOC sidebar (260px)
- `frontend/src/lib/mobile-layout.svelte.ts` — mobile state (breakpoint at 1024px)
- `frontend/src/lib/theme.css` — CSS variables

## Checklist

- [x] Create a `ResizeHandle.svelte` component with drag behavior (mousedown → mousemove → mouseup), visual grab handle, and cursor feedback
- [x] Add resize handle between left sidebar and center content in `+page.svelte`; drive sidebar width from reactive state instead of fixed CSS
- [x] Add resize handle between center content and right chat panel; drive chat panel width from reactive state
- [x] Enforce min/max width constraints (sidebar: 180–640px, chat: 240–800px)
- [x] Persist panel widths to localStorage and restore on load
- [x] Ensure resize handles are hidden on mobile (< 1024px) and panels revert to fixed/overlay behavior

## Notes

- The resize handle should be a thin (4–6px) invisible hit area overlaid on the panel border, with `cursor: col-resize`. A subtle visual indicator (line or dots) can appear on hover.
- Use `pointer events` (pointerdown/pointermove/pointerup) instead of mouse events for better touch/pen support.
- Consider whether the TOC panel inside PdfViewer also needs a resize handle — defer to a follow-up unless trivial to include.
- The center content should never shrink below ~300px regardless of sidebar + chat widths.
