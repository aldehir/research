# Task 06: Fix PDF text selection span geometry

Text layer `<span>` elements extend beyond the visible PDF canvas, making text selection inaccurate. The span geometry is ~33% wider than the rendered content due to a scale mismatch between the viewport and the CSS `--total-scale-factor`.

## Context

**Root cause:** In `pdf-render.ts`, the viewport is created with `scale: currentScale`, but `--total-scale-factor` is set to `currentScale * PDF_TO_CSS_UNITS` (where `PDF_TO_CSS_UNITS = 96/72 ≈ 1.333`). The pdf.js TextLayer CSS uses `--total-scale-factor` to compute font sizes, but span positions are percentages of the unscaled page dimensions rendered into a container sized from the viewport (which lacks the CSS units factor). This mismatch makes spans ~33% too wide.

**How pdf.js TextLayer sizing works:**
- Span positions: `left: (100 * x / pageWidth)%` — percentage of unscaled page
- Span font-size: `calc(var(--total-scale-factor) * var(--min-font-size) * var(--font-height))`
- Span width correction: `transform: scaleX(var(--scale-x))` where `--scale-x` is computed from canvas text measurement
- Container width: `viewport.width = pageWidth * currentScale` (missing `PDF_TO_CSS_UNITS`)

**The standard pdf.js PDFViewer** creates viewports with `scale: currentScale * PDF_TO_CSS_UNITS`, making container dimensions match the CSS scaling. Our standalone usage doesn't do this.

**Key files:**
- `frontend/src/lib/pdf-render.ts` — `renderPage()` creates viewport and sets `--total-scale-factor`
- `frontend/src/lib/PdfViewer.svelte` — uses `getPageDimensions()` for placeholder sizing, `computeFitScale()` for initial scale
- `frontend/src/lib/pdf-utils.ts` — `fitToWidthScale()` computes scale from container width
- `frontend/node_modules/pdfjs-dist/web/pdf_viewer.css` lines 916-930 — TextLayer CSS rules

## Checklist

- [x] Reproduce and verify the span geometry mismatch (inspect spans at various zoom levels)
- [x] Fix the viewport/container scale to include `PDF_TO_CSS_UNITS` so container dimensions match `--total-scale-factor`
- [x] Adjust `getPageDimensions()` and `fitToWidthScale()` callers to account for the corrected scale
- [x] Verify text selection aligns with visible text at multiple zoom levels and DPI settings

## Notes

- The fix likely involves creating the viewport with `scale: currentScale * PDF_TO_CSS_UNITS` in `renderPage()`, then adjusting `getPageDimensions()` similarly. The canvas already handles DPI separately via `devicePixelRatio`.
- `fitToWidthScale()` computes scale from `containerWidth / pageWidth` — this may need to divide by `PDF_TO_CSS_UNITS` so the user-facing scale value remains intuitive (1.0 = 100%).
- Task 04 (PDF text selection highlight misalignment) was a prior fix that addressed a different aspect of the same area — the `--total-scale-factor` CSS variable was added then.
