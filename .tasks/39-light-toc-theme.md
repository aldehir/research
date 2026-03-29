# Task 39: Light theme for ToC panel

In light mode, the ToC panel uses dark background/text colors (same as the header), making it look like a dark-mode panel. Update the light theme ToC CSS variables so the panel matches the overall light aesthetic.

## Context

- Theme variables are in `frontend/src/lib/theme.css`
- ToC component is `frontend/src/lib/TocPanel.svelte` — uses `--color-toc-bg`, `--color-toc-text`, `--color-toc-heading`, `--color-toc-border`, `--color-toc-active`, `--color-toc-hover`
- Light mode ToC currently uses dark values:
  - `--color-toc-bg: oklch(0.2795 0.0369 290)` — very dark (matches `--color-bg-invert`)
  - `--color-toc-text: oklch(0.869 0.0199 283)` — light text for dark bg
  - `--color-toc-heading: oklch(0.9288 0.0127 286)` — light heading for dark bg
  - `--color-toc-border: oklch(0.3717 0.0392 287)` — dark border
  - `--color-toc-active: oklch(0.6231 0.1881 290 / 0.15)` — purple tint on dark
  - `--color-toc-hover: oklch(1 0 0 / 0.08)` — white overlay on dark
- Dark mode ToC values are already correct and should not change
- The `TocPanel.svelte` component only references `--color-toc-*` and `--color-text-tertiary` variables, so only theme.css needs updating

## Checklist

- [x] Update light theme `--color-toc-*` variables in `theme.css` to use light background, dark text, and matching border/active/hover states
- [x] Verify dark theme ToC variables remain unchanged
- [ ] Visual check in both themes

## Notes

- Use existing light theme palette tokens (e.g. `--color-bg-secondary` range, `--color-text` range) as reference for appropriate lightness values
- Keep the purple hue angle (~286-293) consistent with the rest of the theme
- The `toc-toggle` button uses `--color-text-tertiary` (not a toc-specific variable), which already adapts correctly per theme
