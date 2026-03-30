# Task 56: Fix horizontal scrolling when zoomed into PDF

Horizontal scrolling no longer works when the PDF is zoomed in past the viewport width.

## Context

The PDF viewer (`frontend/src/lib/PdfViewer.svelte`) uses a flex column layout with `overflow: auto` on `.pages-container`. When zoomed in, pages become wider than the container, which should enable horizontal scrolling.

Task 44 previously fixed horizontal panning by changing `align-items: center` to `align-items: flex-start` and adding `margin-inline: auto` to `.page-wrapper` — both changes are still in place (lines 740, 753).

Key code areas:
- `handleWheel()` (line 417): Only intercepts Ctrl/Cmd+scroll for zoom; returns early for normal scroll without calling `preventDefault()` — this looks correct
- Conditional wheel binding (line 609): `onwheel={selectionMode ? undefined : handleWheel}` — handler is always attached in normal mode
- `.pages-container` CSS (line 733): `overflow: auto`, `display: flex`, `flex-direction: column`, `align-items: flex-start`
- `.page-wrapper` CSS (line 749): `margin-inline: auto`, `flex-shrink: 0`
- `.pages-area` (line 727): `flex: 1`, `position: relative`, `min-height: 0` — wraps both scroll container and region select overlay

Possible causes to investigate:
- The `.pages-area` wrapper may be constraining the scroll container width (no `min-width: 0` or `overflow: hidden` interactions)
- The flex column layout may be collapsing horizontal overflow in some browsers
- A recent change may have altered the container sizing chain

## Checklist

- [x] Reproduce the issue — zoom in and verify horizontal scroll is broken
- [x] Identify root cause in CSS/layout chain
- [x] Fix the layout so horizontal scrolling works when pages overflow
- [x] Verify fix doesn't break centered pages at fit-to-width zoom

## Notes

- Task 44 (commit d2a5f8b) fixed the same area before — review its approach as baseline
- The `.pages-area` div was added later (region select overlay refactor, commit 94ab9d5) and may have introduced the regression
