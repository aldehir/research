# Task 22: Overlay delete icon in paper side menu

Make the delete icon in the paper list sidebar overlay the paper item so it remains visible even when the paper title overflows.

## Context

- **Component**: `frontend/src/lib/PaperList.svelte` renders each paper as an `<li>` with a `.paper-item` button (flex column: title + meta) and a `.delete-btn` sibling
- **Current layout**: `li` is `display: flex; align-items: center` — the delete button sits beside the paper-item as a flex sibling
- **Title overflow**: `.paper-title` uses `overflow: hidden; text-overflow: ellipsis; white-space: nowrap`
- **Icons**: `Icon.svelte` + `X` path from `$lib/icons`
- **Theme**: uses CSS variables from `theme.css` (`--color-danger`, `--color-surface-hover`, `--color-surface-active`, `--color-bg-secondary`)

## Checklist

- [x] Position delete button with `position: absolute` on the right side of the `li` (make `li` `position: relative`)
- [x] Add a gradient or solid background to the delete button so it doesn't clash with the truncated title text beneath it
- [x] Ensure the overlay respects hover/selected background states (surface-hover, surface-active)
- [x] Verify mobile touch target size is preserved (44px min)

## Notes

- The gradient background on the delete button should blend with the item's current background state (default, hover, selected) to avoid a visible hard edge.
- Consider using `transparent → background-color` gradient on a pseudo-element or padding area to the left of the icon.
