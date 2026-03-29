# Task 41: Fix theme toggle requiring double click

The theme toggle button in the header requires two clicks to visually swap. The CSS theme applies on the first click (via `data-theme` attribute change), but the Svelte `{#if}` block controlling the Sun/Moon icon does not re-render, making it appear that the toggle didn't work.

## Context

- **State management**: `frontend/src/lib/theme.svelte.ts` — module-level `$state<Theme>('light')` with getter/setter functions (`getTheme`, `setTheme`)
- **Toggle button**: `frontend/src/routes/+layout.svelte:134-145` — uses `{#if getTheme() === 'light'}` to pick Sun/Moon icon
- **Cycle function**: `frontend/src/routes/+layout.svelte:101-103` — `cycleTheme()` reads `getTheme()` and calls `setTheme()` with the opposite value
- **DOM application**: `applyTheme()` sets `data-theme` attribute on `document.documentElement`, which swaps CSS variables immediately

The likely root cause is that the `{#if getTheme() === 'light'}` template expression doesn't properly track the `$state` signal read inside `getTheme()`. The CSS swap works because `applyTheme` directly mutates the DOM, but the Svelte template block doesn't re-evaluate because the function call boundary may prevent the compiler from tracking the signal dependency.

## Checklist

- [x] Write a test verifying theme state toggles correctly on a single call
- [x] Fix reactivity so the template re-renders on theme change (e.g., use `$derived` in component or export reactive state directly)
- [x] Verify icon swaps immediately on a single click

## Notes

- Svelte 5 runes mode: `$state` in `.svelte.ts` modules should be reactive when read in components, but accessing through a plain function call may not trigger template updates in `{#if}` blocks
- The `app.html` inline script separately handles initial theme to prevent FOUC — that code path is unrelated to this bug
- The `data-theme` attribute change (CSS swap) works fine; only the Svelte template reactivity is broken
