# Task 09: Add table of contents with page jumping

Allow users to view a PDF's table of contents (outline/bookmarks) and click entries to jump to the corresponding page.

## Context

Summary of relevant existing code:

- **pdf.js `getOutline()` API** — already loaded via `pdfjs-dist`; returns outline tree with `dest` properties that resolve to page numbers via `pdfDoc.getDestination()` + `pdfDoc.getPageIndex()`
- **`PdfViewer.svelte`** — main viewer component (567 lines); already has `goToPage()` for page jumping and a toolbar with page/zoom controls
- **`+page.svelte`** — 3-panel layout: left sidebar (paper list, 280px), center (PDF viewer), right (chat panel, 360px collapsible). ChatPanel provides a collapsible sidebar pattern to follow.
- **`pdf-render.ts`** / **`pdf-utils.ts`** — rendering and zoom utilities
- **No backend PDF parsing** exists today; outline extraction should happen client-side via pdf.js since the library is already loaded and has full outline support. No Go PDF library or schema changes needed.

## Checklist

- [x] Extract outline from loaded PDF document using `pdfDoc.getOutline()` and resolve destinations to page numbers
- [x] Create `TocPanel.svelte` component that renders the outline as a nested tree with expand/collapse for sub-items
- [x] Wire TOC entry clicks to `goToPage()` in PdfViewer for page jumping
- [x] Add a toolbar button in PdfViewer to toggle TOC panel visibility
- [x] Handle PDFs with no outline (hide button or show empty state)
- [x] Style TOC panel consistent with existing sidebar patterns (dark theme, scrollable)

## Notes

- Client-side only — pdf.js `getOutline()` returns the full bookmark tree without needing backend changes
- Outline entries can have nested children (recursive tree structure)
- Some PDFs have no outline at all; the UI should handle this gracefully
- Destinations can be named or explicit; need to handle both via `pdfDoc.getDestination()` and `pdfDoc.getPageIndex()`
- Consider highlighting the current TOC entry based on `currentPage` scroll position
