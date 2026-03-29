# Task 36: Mobile panel layout tweaks

Adjust mobile overlay panels: remove border-radius from the paper sidebar panel so it matches the chat panel, and increase the chat panel width to 75vw.

## Context

Summary of relevant existing code discovered during exploration:
- Mobile layout is triggered via `window.matchMedia('(max-width: 1023px)')` in `frontend/src/routes/+layout.svelte`
- Mobile panel state managed in `frontend/src/lib/mobile-layout.svelte.ts`
- Sidebar mobile overlay: fixed 280px, slides from left (`+layout.svelte:445-449`)
- Chat mobile overlay wrapper: 70vw / max 400px / min 280px, slides from right (`+layout.svelte:452-460`)
- Neither panel has explicit `border-radius` on the overlay container, but inner elements (buttons, drop overlay) use `var(--radius)` — the drop-overlay at line 357 has `border-radius: var(--radius)` which may be visible inside the sidebar
- Chat panel has no border-radius on its outer container either
- The user observes a visual inconsistency between the two panels regarding border-radius

## Checklist

- [x] Identify and remove border-radius causing visual inconsistency on the paper sidebar in mobile overlay mode
- [x] Increase chat overlay wrapper width from 70vw to 75vw
- [x] Verify both panels look consistent in mobile layout

## Notes

- The border-radius discrepancy may come from the `.drop-overlay` inside the sidebar or from inherited styles — needs visual inspection or careful CSS audit
- Chat panel max-width (400px) and min-width (280px) constraints may also need adjustment to complement the 75vw change
