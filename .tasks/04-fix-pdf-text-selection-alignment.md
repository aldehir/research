# Task 04: Fix PDF text selection highlight misalignment

The visual text selection highlight on the PDF does not align with the actual text. The selection either extends beyond the PDF boundary or doesn't reach the text edges.

## Context

Key file: `frontend/src/lib/PdfViewer.svelte`

- The component imports `pdfjs-dist/web/pdf_viewer.css` (line 2), which provides default `.textLayer` styling including positioning and text span transforms.
- Custom CSS at lines 360-368 overrides `.textLayer` with `position: absolute; top/left/right/bottom: 0; overflow: hidden; line-height: 1;`. This likely **conflicts** with the styles from `pdf_viewer.css`, which already handles text layer positioning correctly.
- The `renderPage` function (line 77) creates a TextLayer with the viewport but doesn't explicitly set the text layer div dimensions — it relies on CSS to size/position it.
- Canvas uses `devicePixelRatio` scaling (lines 90-99) for hi-DPI rendering, but the text layer receives the same viewport without any DPI compensation.

Root cause hypothesis: The custom `.textLayer` CSS override conflicts with `pdf_viewer.css`'s built-in text layer positioning (which uses CSS variables and transforms on individual spans). The `line-height: 1` and blanket absolute positioning likely break the precise span positioning that pdf.js calculates.

## Checklist

- [x] Investigate the exact styles from `pdfjs-dist/web/pdf_viewer.css` that apply to `.textLayer` and its children
- [x] Write a test verifying text layer render produces correct structure (textLayer div with spans)
- [x] Remove or reduce custom `.textLayer` CSS overrides that conflict with `pdf_viewer.css`
- [x] Verify text selection alignment on standard and hi-DPI displays
- [x] Test that selection still triggers `handleMouseUp` and populates `setSelection` correctly

## Notes

- `pdfjs-dist/web/pdf_viewer.css` is the canonical stylesheet for pdf.js text layers; custom overrides should be minimal.
- The `line-height: 1` override is suspicious — pdf.js sets its own line-height per span based on font metrics.
- May need to ensure the text layer div gets explicit `width`/`height` matching the viewport, rather than relying on `right: 0; bottom: 0`.
