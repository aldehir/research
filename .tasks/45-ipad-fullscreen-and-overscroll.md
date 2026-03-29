# Task 45: iPad fullscreen mode and overscroll bounce fix

Maximize screen real estate on iPad by adding a fullscreen mode that hides the header, and prevent the Safari rubber-band bounce effect when content isn't scrollable.

## Context

- Layout lives in `frontend/src/routes/+layout.svelte` — three-column flexbox with a 48px header (`.app-header`)
- `.app-shell` uses `height: 100vh`; `.content` has `overflow: hidden`
- No `overscroll-behavior` properties exist anywhere in the codebase
- Mobile breakpoint is 1024px — iPad is treated as desktop in landscape, mobile in portrait
- Viewport meta in `app.html`: standard `width=device-width, initial-scale=1`
- Touch-aware patterns already exist in `ResizeHandle.svelte` (`@media (pointer: coarse)`, `touch-action: none`)

## Checklist

- [x] Add `overscroll-behavior: none` to html/body and scroll containers to prevent Safari rubber-band bounce
- [x] Add viewport meta `viewport-fit=cover` for proper iPad fullscreen support
- [x] Add fullscreen toggle button to the header that hides the header bar (reclaims 48px)
- [x] Persist fullscreen preference (localStorage)
- [x] Ensure panels, PDF viewer, and chat adapt to the freed vertical space when header is hidden
- [x] Add a way to exit fullscreen (e.g. swipe down from top edge or a small floating button)

## Notes

- Safari's overscroll bounce happens on the `<html>`/`<body>` level when no element is scrollable, or at the edges of scroll containers. `overscroll-behavior: none` on `html, body` should suppress it. Also consider `overflow: hidden` on `html` since the app is a fixed-layout SPA.
- For fullscreen, hiding the 48px header is the biggest win. Could also consider hiding the browser chrome via `minimal-ui` or PWA manifest, but that's a separate concern.
- The fullscreen exit affordance needs to be discoverable but not intrusive — a small semi-transparent pill at the top edge could work.
