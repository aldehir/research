# Task 42: Reduce saturation of purple colors in dark theme

Desaturate the purple tones in the dark theme for a more muted, comfortable appearance.

## Context

All theme colors live in `frontend/src/lib/theme.css` using OKLCH color space. The dark theme (`[data-theme="dark"]`) uses purple hues (280–296) throughout. The chroma (second OKLCH value) controls saturation.

High-chroma purple colors in dark theme:
- `--color-primary: oklch(0.6231 0.1881 290)` — main accent
- `--color-primary-hover: oklch(0.5461 0.2153 293)` — accent hover
- `--color-toc-active: oklch(0.6231 0.1881 290 / 0.2)` — ToC active item

Medium-chroma purple colors (background/surface tints):
- `--color-primary-light: oklch(0.3462 0.0736 286)`
- `--color-surface-active: oklch(0.3462 0.0736 286)`
- `--color-surface-hover: oklch(0.3074 0.0487 293)`
- `--color-bg-tertiary: oklch(0.2858 0.0442 293)`
- `--color-bg-secondary: oklch(0.2522 0.0397 294)`
- `--color-header-bg / --color-toc-bg / --color-pages-bg: oklch(0.1963 0.0402 293)`
- `--color-bg: oklch(0.231 0.0366 296)`

## Checklist

- [x] Reduce chroma of high-saturation accent colors (`--color-primary`, `--color-primary-hover`, `--color-toc-active`)
- [x] Reduce chroma of medium-saturation surface/background colors
- [x] Visually verify dark theme still has enough contrast and visual hierarchy

## Notes

- OKLCH chroma is the second value — lower = more muted. Keep hue angles intact.
- Don't touch the light theme.
- Ensure primary accent remains visually distinct from background even after desaturation.
