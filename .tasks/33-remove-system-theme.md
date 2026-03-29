# Task 33: Remove system theme selection

Remove the "system" option from the theme selector, keeping only light and dark. The toggle should cycle between light and dark directly.

## Context

- `frontend/src/lib/theme.svelte.ts` — defines `Theme` type as `'light' | 'dark' | 'system'`, defaults to `'system'`, `applyTheme()` removes `data-theme` attr for system mode, `getResolvedTheme()` checks `prefers-color-scheme` for system
- `frontend/src/routes/+page.svelte` — `cycleTheme()` cycles through `['light', 'system', 'dark']`, renders Sun/Monitor/Moon icons based on current theme
- `frontend/src/lib/theme.css` — has a `@media (prefers-color-scheme: dark)` block (~line 100–140) that duplicates all dark-mode variables for `:root:not([data-theme])` (system fallback)
- `frontend/src/lib/PdfViewer.svelte` — has `:global(:root:not([data-theme]))` selectors for system-mode dark PDF inversion (~lines 733–739)
- Icons: `Monitor` icon is used for system theme; can stop importing it if unused elsewhere

## Checklist

- [x] Update `Theme` type and state in `theme.svelte.ts` — remove `'system'`, default to `'light'`, simplify `applyTheme`/`initTheme`/`getResolvedTheme`
- [x] Update `+page.svelte` — cycle between light/dark only, remove Monitor icon if unused, update toggle rendering
- [x] Remove system preference fallback CSS block from `theme.css`
- [x] Remove `:root:not([data-theme])` selectors from `PdfViewer.svelte`
- [x] Verify no other files reference `'system'` theme or `:root:not([data-theme])`

## Notes

- Users with `theme=system` in localStorage will need migration — `initTheme` should treat unrecognized/`'system'` values as `'light'` (or could resolve system preference one final time during migration).
- The `Monitor` icon export in `$lib/icons/index.ts` can stay unless cleanup is desired — it's harmless.
