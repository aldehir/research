# Task 18: Add subpixel antialiasing to PDF rendering

Improve PDF text rendering sharpness by enabling subpixel antialiasing on the canvas. Without an explicit background color on the canvas element, browsers fall back to grayscale antialiasing, producing blurry text. Mozilla's pdf.js demo viewer addresses this with custom canvas CSS and rendering options.

## Context

Summary of relevant existing code:
- **`frontend/src/lib/pdf-render.ts`** — `renderPage()` creates a canvas, scales it by `devicePixelRatio`, and calls `page.render()`. Currently no background color is set on the canvas element itself, only on the `.page-wrapper` container.
- **`frontend/src/lib/PdfViewer.svelte`** — Main component with CSS styles. `.page-wrapper` has `background: white` but the `<canvas>` inside it does not.
- Canvas rendering at lines 76-89 of `pdf-render.ts`: creates canvas, sets DPI scaling, renders page — no `canvas.style.backgroundColor` or `canvasContext` background fill.
- Mozilla's pdf.js demo viewer sets `background-color: white` on the canvas CSS and uses `pageColors` / background fill to enable the browser's subpixel text rendering path.

## Checklist

- [x] Add `background-color: white` CSS to canvas elements in the page wrapper
- [x] Pre-fill canvas with white before rendering to ensure opaque background for compositing
- [x] Verify rendering improvement visually on LCD display at various zoom levels
- [x] Check that transparent PDF elements (if any) still render correctly over white background

## Notes

- Browsers only use subpixel (LCD) antialiasing when they know the background color at compositing time. A transparent canvas forces grayscale AA.
- The fix is purely frontend CSS/canvas — no backend changes needed.
- Mozilla's viewer also sets `-webkit-font-smoothing: subpixel-antialiased` on some elements, worth testing.
- The `page.render()` call accepts a `background` option in some pdf.js versions — check if pdfjs-dist v5.5 supports it.
