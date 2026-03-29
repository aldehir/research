# Task 35: Hideable sidebar panel in desktop view

Add a collapse/expand toggle to the sidebar (paper list) panel in desktop view, matching the pattern already used by the chat panel.

## Context

- `frontend/src/lib/ChatPanel.svelte` — already implements a collapse pattern: a `collapsed` $state boolean, renders `.chat-collapsed` button when true, full panel when false. Uses `PanelRightOpen`/`PanelRightClose` icons. Collapse toggle only shown on desktop (`{#if !getIsMobile()}`).
- `frontend/src/routes/+page.svelte` — desktop layout renders sidebar always-visible in a flex row: `aside.sidebar` (width from `getSidebarWidth()`) → `ResizeHandle` → center (PDF viewer) → `ResizeHandle` → `ChatPanel`. Mobile uses overlay pattern with `activePanel` state.
- `frontend/src/lib/panel-widths.svelte.ts` — manages `sidebarWidth`/`chatWidth` with localStorage persistence and getters/setters.
- `frontend/src/lib/mobile-layout.svelte.ts` — mobile panel toggling with `activePanel` state. Not relevant for desktop collapse.
- `frontend/src/lib/icons/index.ts` — has `PanelRightOpen`, `PanelRightClose` for the chat collapse. Will need `PanelLeftOpen`/`PanelLeftClose` (or similar) for sidebar.
- Tests: `tests/mobile-layout.test.ts` covers mobile toggling; `tests/panel-widths.test.ts` covers width management. No existing tests for chat collapse state.

## Checklist

- [x] Add `PanelLeftOpen` and `PanelLeftClose` icon paths to `$lib/icons/index.ts`
- [x] Add sidebar collapse state and toggle — either in `+page.svelte` or a new store function in `panel-widths.svelte.ts`
- [x] Render collapsed sidebar as a narrow button strip (matching chat collapse pattern) and hide `ResizeHandle` when collapsed
- [x] Ensure sidebar width restores to previous value when expanding
- [x] Collapse toggle only visible on desktop (not mobile)

## Notes

- Follow the ChatPanel.svelte collapse pattern closely for consistency.
- The sidebar is rendered directly in `+page.svelte` (not a separate component like ChatPanel), so the collapse state and rendering will live there.
- When collapsed, the resize handle between sidebar and center should be hidden.
- Consider keyboard shortcut later (not in scope).
