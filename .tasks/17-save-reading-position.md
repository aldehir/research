# Task 17: Save reading position per paper

Persist the current page number for each paper so the viewer reopens where the user left off. The frontend sends position updates via a debounced API call to avoid excessive writes during scroll.

## Context

Summary of relevant existing code:

- **`internal/store/db.go`** — SQLite schema with migrations; papers table has no position column yet
- **`internal/store/papers.go`** — `Paper` struct with `PageCount *int` field; CRUD functions using raw SQL
- **`internal/api/papers.go`** — Handlers for `GET/POST/DELETE /api/papers/{id}`; no PATCH endpoint exists
- **`frontend/src/lib/PdfViewer.svelte`** — Tracks `currentPage` via `$state(1)`, updated in `handleScroll()` which computes closest page to viewport center
- **`frontend/src/lib/pdf-context.svelte.ts`** — Shared `currentPage` state with `setCurrentPage`/`getCurrentPage`
- **`frontend/src/lib/api.ts`** — API client; `Paper` interface has no position field; `MessageContext` already sends `currentPage` per chat message but it's not persisted

Patterns to follow: raw SQL migrations via `ALTER TABLE`, `stretchr/testify` for Go tests, Svelte 5 runes for state, debounce for scroll-driven updates.

## Checklist

- [ ] Add `last_read_page` column to papers table (migration in `db.go`) and update `Paper` struct
- [ ] Add `UpdateReadingPosition(ctx, paperID, page)` store function with test
- [ ] Return `last_read_page` in `GetPaper` / `ListPapers` responses (verify with test)
- [ ] Add `PATCH /api/papers/{id}/position` handler with test
- [ ] Add `updateReadingPosition(id, page)` to frontend API client and extend `Paper` type
- [ ] Debounced save on scroll — call API when `currentPage` changes, debounce ~2s
- [ ] Restore position on PDF load — read `last_read_page` from paper and `goToPage` on mount

## Notes

- Debounce strategy: fire API call ~2s after last `currentPage` change. This avoids a request per scroll frame while still saving frequently enough to be useful.
- Only save when the page actually changes (deduplicate consecutive identical values).
- No need for a `reading_position_updated_at` column unless we find a use for it later.
- Consider using `navigator.sendBeacon` or a `beforeunload` handler as a fallback to save position when the tab closes mid-debounce.
