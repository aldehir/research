# Task 28: Dark mode PDF color inversion

Invert PDF page colors in dark mode so the rendered pages blend with the dark theme instead of showing bright white. Use CSS filters (`invert`, `hue-rotate`, `brightness`, `contrast`) to soften the result so it aligns with the dark color scheme rather than producing harsh pure black.

## Context

Summary of relevant existing code:

- **`frontend/src/lib/PdfViewer.svelte`** — renders PDF pages in `.page-wrapper` divs containing canvases. Canvas background is hardcoded white (line 720). Page wrapper background is also hardcoded white (line 714).
- **`frontend/src/lib/pdf-render.ts`** — `renderPage()` sets `canvas.style.backgroundColor = 'white'` (line 82) and pre-fills canvas with `#ffffff` for subpixel antialiasing (line 91).
- **`frontend/src/lib/theme.css`** — defines dark theme under `[data-theme="dark"]` and system fallback under `@media (prefers-color-scheme: dark) :root:not([data-theme])`. Dark theme background is `--color-bg: #1a1b2e`.
- **`frontend/src/lib/theme.svelte.ts`** — theme store with `getResolvedTheme()` returning `'light' | 'dark'`.

Approach: pure CSS using `filter: invert(1) hue-rotate(180deg) brightness() contrast()` on the page-wrapper canvas in dark mode. Tune brightness/contrast so inverted black becomes a soft dark gray close to `--color-bg` rather than pure black. Apply to both `[data-theme="dark"]` and the system preference media query fallback.

## Checklist

- [x] Add dark-mode CSS filter rules to `.page-wrapper :global(canvas)` in PdfViewer.svelte for `[data-theme="dark"]`
- [x] Add matching rules for `@media (prefers-color-scheme: dark)` system fallback
- [x] Update `.page-wrapper` background from hardcoded white to theme-aware (white in light, dark in dark)
- [x] Tune brightness/contrast values so the result looks comfortable against `--color-bg: #1a1b2e`
- [x] Verify text layer and annotation layer remain functional on top of inverted canvas

## Notes

- `hue-rotate(180deg)` paired with `invert(1)` preserves color semantics in figures/charts
- Images/photos rendered on the PDF canvas will appear inverted — no workaround since pdf.js renders everything to a single canvas
- The subpixel antialiasing pre-fill in `pdf-render.ts` (white fillRect) should stay as-is; the CSS filter operates on the composited result
- Start with something like `brightness(0.85) contrast(0.85)` and adjust by eye
