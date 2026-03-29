# Task 30: Convert CSS colors to OKLCH

Convert all hex and rgba color values to OKLCH color space for perceptually uniform color representation. Create a Python conversion script to automate the transformation.

## Context

Summary of relevant existing code:
- `frontend/src/lib/theme.css` — All CSS custom properties defined here across 3 blocks: `:root` (light), `[data-theme="dark"]`, and `@media (prefers-color-scheme: dark)` fallback. Uses hex colors (`#ffffff`, `#2563eb`, etc.) and `rgba()` for semi-transparent values.
- `frontend/src/lib/TocPanel.svelte:164` — hardcoded `rgba(255, 255, 255, 0.08)`
- `frontend/src/routes/+page.svelte:301,420` — hardcoded `rgba(255, 255, 255, 0.1)`
- No hex colors are hardcoded in `.svelte` components (only in `theme.css`).
- Dark theme block and system-preference fallback block have identical values — keep them in sync.

## Checklist

- [x] Create Python script (`scripts/hex-to-oklch.py`) that converts hex and rgba to oklch format
- [x] Convert all `:root` (light theme) colors in `theme.css` to oklch
- [x] Convert all `[data-theme="dark"]` colors in `theme.css` to oklch
- [x] Convert all `@media (prefers-color-scheme: dark)` fallback colors to oklch (must match dark theme)
- [x] Convert hardcoded `rgba()` values in `TocPanel.svelte`, `+page.svelte`, and `MarkdownRenderer.svelte` to theme variables
- [x] Visually verify no color regressions (light and dark modes)

## Notes

- OKLCH format: `oklch(L C H)` where L=lightness (0-1), C=chroma (0-0.4+), H=hue (0-360)
- For semi-transparent colors, use `oklch(L C H / alpha)` syntax
- Pure white (`#ffffff`) = `oklch(1 0 0)`, pure black (`#000000`) = `oklch(0 0 0)`
- The Python script is a one-time utility but kept in `scripts/` for future use
- `rgba(0, 0, 0, alpha)` values like `--color-shadow` and `--color-backdrop` become `oklch(0 0 0 / alpha)`
