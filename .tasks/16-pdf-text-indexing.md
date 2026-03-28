# Task 16: Background PDF Text Extraction and Search Index

Pre-extract PDF text into SQLite on a background schedule so search and page reads don't shell out to `pdftotext` on every request. Use SQLite FTS5 for ranked full-text search.

## Context

Currently `SearchText` and `ExtractPageText` run `pdftotext` on every call — no caching. This is slow for large PDFs and redundant across searches.

Relevant existing code:
- `internal/pdf/text.go` — `ExtractPageText` and `SearchText` shell out to `pdftotext`
- `internal/pdf/storage.go` — PDFs stored at `{PDF_DIR}/{id}.pdf`
- `internal/store/db.go` — SQLite schema with migration system
- `internal/api/messages.go` — `executeToolCall` calls `pdf.SearchText`/`pdf.ExtractPageText`
- `cmd/server/main.go` — server startup, no background job infra yet

Design decisions:
- **Storage**: `paper_pages` table with `(paper_id, page_num, text)` + FTS5 virtual table for search
- **Extraction trigger**: Background goroutine polls for papers missing extracted text (resilient to upload failures, restarts, etc.)
- **Search**: FTS5 with BM25 ranking replaces naive substring search — gives relevance-ranked results with page numbers for free
- **Read page**: Reads from `paper_pages` table, falls back to `pdftotext` if not yet extracted

## Checklist

- [x] Add `paper_pages` table migration: `(id, paper_id, page_num, text_content, UNIQUE(paper_id, page_num))`
- [x] Add FTS5 virtual table: `paper_pages_fts` using content from `paper_pages`
- [x] Add store functions: `UpsertPageText`, `GetPageText`, `SearchPageText` (FTS5 BM25 query)
- [x] Test store functions with FTS5 search ranking and page-scoped reads
- [x] Add `internal/pdf/indexer.go` — `Indexer` struct that extracts all pages for a paper via `pdftotext` and writes to `paper_pages`
- [x] Test indexer: extracts pages, writes to DB, skips already-indexed papers
- [x] Add background worker in `cmd/server/main.go` — polls for unindexed papers on a ticker, runs indexer
- [x] Wire `executeToolCall` to read from `paper_pages` for `read_page` (fall back to pdftotext)
- [x] Wire `executeToolCall` to use FTS5 search for `search_pdf` (fall back to pdftotext)
- [x] Add `paper_pages` cleanup on paper deletion (CASCADE or explicit)

## Notes

- SQLite FTS5 is built into `modernc.org/sqlite` — no new dependency needed.
- `pdftotext` already separates pages with `\f` — extract once, split, insert per page.
- Background poll interval ~30s is fine. Could also trigger immediately on upload as an optimization later.
- FTS5 `rank` function gives BM25 scores out of the box: `SELECT * FROM paper_pages_fts WHERE paper_pages_fts MATCH ? ORDER BY rank`.
- Keep `pdftotext` fallback for papers not yet indexed — graceful degradation.
- Consider adding a `text_indexed_at` column to `papers` to track extraction status without a separate status enum.
