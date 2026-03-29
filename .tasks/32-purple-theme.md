# Task 32: Shift theme hues from blue to purple

Change the color palette in `theme.css` to be more purple/violet and less blue.

## Context

All theme colors live in `frontend/src/lib/theme.css` using OKLCH color space. Three blocks define colors: `:root` (light), `[data-theme="dark"]`, and `@media (prefers-color-scheme: dark)` fallback.

Current hue values cluster around 248–265 (blue). Purple/violet sits around 280–310 in OKLCH. The shift involves bumping hue values on UI chrome colors (backgrounds, borders, text tints, primary accent, surfaces, header, TOC, pages) while leaving semantically independent colors (danger red, code syntax highlighting) alone.

Key variables to shift:
- `--color-bg-*` (secondary/tertiary/invert) — hues ~248–260
- `--color-text-*` — hues ~255–265
- `--color-border-*` — hues ~252–257
- `--color-primary-*` — hues ~255–262
- `--color-surface-*` — hues ~256–278
- `--color-header-*`, `--color-toc-*`, `--color-pages-bg` — hues ~252–278

Code block chrome (`--color-code-bg`, `--color-code-text`, `--color-code-comment`) can shift slightly since they're already near 272–283, but syntax token colors (keyword, string, function, etc.) should stay as-is.

## Checklist

- [x] Shift light theme (`:root`) hues toward purple (~280–300 range)
- [x] Shift dark theme (`[data-theme="dark"]`) hues to match
- [x] Shift system-preference dark fallback to match `[data-theme="dark"]`
- [x] Visual review — verify no contrast/readability regressions

## Notes

- Keep chroma and lightness values the same or very close; only rotate hue
- Danger/red colors stay unchanged
- Syntax highlighting token colors stay unchanged
- The three dark-mode blocks (explicit + media query fallback) must stay in sync
