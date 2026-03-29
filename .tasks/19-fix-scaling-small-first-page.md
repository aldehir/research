# Task 19: Fix scaling when first page is smaller than others

Fit-to-width scale is computed from the first page's width only. When the first page is narrower than subsequent pages (e.g. cover pages, half-title pages), the uniform scale is too large — later pages overflow the container horizontally.

## Context

- `PdfViewer.svelte` lines 239-243 and `computeFitScale()` (lines 105-117) both use `pages[0]` width to compute fit-to-width scale.
- `pdf-utils.ts` `fitToWidthScale()` takes a single `pageWidth` argument.
- The calculated scale is applied uniformly to all pages via `getPageDimensions()` and `renderPage()`.
- The ResizeObserver (lines 514-533) also recalculates using the same first-page logic.

Fix: use the **maximum** page width across all pages (or a representative sample) when computing fit-to-width scale so that no page overflows.

## Checklist

- [x] Add test for `fitToWidthScale` with varying page widths (returns scale fitting the widest page)
- [x] Update `fitToWidthScale` or add helper to accept multiple page widths
- [x] Update `PdfViewer.svelte` initial load to pass widest page width
- [x] Update `computeFitScale()` to use widest page width
- [x] Update ResizeObserver handler to use widest page width
- [ ] Manual test with a PDF whose first page is narrower than later pages

## Notes

- All page objects are already loaded (`allPages` / `pages` array) before scale is computed, so iterating widths is cheap (just `getViewport({scale:1})` per page — no rendering).
- Only the width matters for fit-to-width; heights vary naturally and are handled by scroll.
- Consider caching the max intrinsic width to avoid recomputing on every resize.
