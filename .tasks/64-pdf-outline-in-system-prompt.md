# Task 64: Include PDF outline in system prompt

Send the PDF document outline (table of contents / bookmarks) in the LLM system prompt so the model knows the document structure and where to look.

## Context

- The frontend already extracts PDF outlines client-side via pdf.js (`frontend/src/lib/pdf-outline.ts`) for the ToC panel, but the backend has no outline extraction
- The system prompt is built in `internal/chat/prompt.go` via `BuildSystemPrompt(PromptContext{...})` — currently includes title, author, date, and page count only
- `internal/pdf/metadata.go` uses `ledongthuc/pdf` to extract metadata at upload time (`internal/api/papers.go:102-125`)
- The outline should be stored in the `papers` table (new `outline_json` column) so it doesn't need to be re-extracted on every chat message
- PDFs without an outline should simply omit the section from the prompt

## Plan

Extract the PDF outline tree using the `ledongthuc/pdf` library (already a dependency), store it as JSON on the papers table during upload alongside other metadata, and include it in the system prompt as an indented list with page numbers.

## Checklist

- [x] Add `ExtractOutline()` to `internal/pdf/outline.go` — walk the PDF Outlines dict from the document catalog, return `[]OutlineEntry{Title, Page, Children}` (test with a PDF that has bookmarks)
- [x] Add `outline_json TEXT` column to papers table (new migration in `internal/store/db.go`)
- [x] Add `OutlineJSON *string` field to `store.Paper`, update all SELECT queries (`GetPaper`, `ListPapers`) and `CreatePaper`
- [x] Store extracted outline as JSON during upload in `internal/api/papers.go` alongside other metadata
- [x] Add `Outline` field to `chat.PromptContext`, format as indented list in `BuildSystemPrompt()` — omit section if empty
- [x] Wire up in message handler (`internal/api/messages.go:196-220`) — load outline from paper record, pass to prompt context

## Notes

- The `ledongthuc/pdf` library provides raw PDF object access via `r.Trailer()` — the outline tree lives at `Catalog → Outlines → First/Last/Next` following the PDF spec's bookmark dictionary chain
- If `ledongthuc/pdf` outline traversal is too fiddly, fallback option is shelling out to `mutool show` (mupdf-tools)
- Existing papers in the DB won't have outlines — the column is nullable, handle gracefully
