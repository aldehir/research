# Task 26: Upload button in sidebar header and full-panel drop zone

Add an upload button next to the "Papers" title in the sidebar header, remove the dedicated UploadZone at the bottom, and make the entire sidebar panel a drag-and-drop area for PDF uploads.

## Context

Current sidebar layout in `+page.svelte`:
```
┌──────────────────────┐
│ Papers               │  ← .sidebar-header (plain text, nothing on the right)
├──────────────────────┤
│ [Paper 1]          ✕ │
│ [Paper 2]          ✕ │  ← PaperList (flex: 1, scrollable)
├──────────────────────┤
│ Drop PDF here or     │
│ click to upload      │  ← UploadZone (separate component at bottom)
└──────────────────────┘
```

Key files:
- `frontend/src/routes/+page.svelte` — sidebar structure: header → `<PaperList />` → `<UploadZone />`
- `frontend/src/lib/UploadZone.svelte` — drag-drop + hidden file input, calls `papersStore.upload(file)`
- `frontend/src/lib/PaperList.svelte` — scrollable paper list with select/delete
- `frontend/src/lib/papers.svelte.ts` — store with `upload(file)` method
- `frontend/src/lib/icons/index.ts` — has `Plus` icon available but unused

Upload flow: file → `papersStore.upload(file)` → `uploadPaper(file)` from api.ts → `POST /api/papers` with FormData.

## Checklist

- [x] Add an upload button (Plus icon) to the right side of `.sidebar-header` next to "Papers" title; clicking it opens a hidden file input for PDF selection
- [x] Move drag-and-drop handling (dragover/dragleave/drop) from UploadZone up to the sidebar container so the entire panel is a drop target, with a visual overlay indicating drop state
- [x] Remove the `<UploadZone />` component from the sidebar layout in `+page.svelte`
- [x] Show a full-panel drag overlay (e.g. dashed border + "Drop PDF" message) when a file is dragged over the sidebar
- [x] Ensure upload error feedback still works (inline error message near header or toast)

## Notes

- The `Plus` icon is already exported from `$lib/icons` but unused — use it for the upload button.
- The UploadZone.svelte file can be deleted or gutted once its logic is absorbed into the sidebar. Prefer inlining the drag-drop logic into `+page.svelte` rather than keeping UploadZone as a wrapper.
- Keep the hidden `<input type="file" accept=".pdf">` pattern for the button click path.
- The drag overlay should cover the entire sidebar (paper list included) with a semi-transparent overlay so it's clear where to drop.
