# Task 63: Update search prompt to use sparse keywords

Update the `search_pdf` tool description to instruct the LLM to use sparse keywords instead of full phrases, since the backend search is keyword-based (FTS5 MATCH).

## Context

- Tool definitions: `internal/chat/tools.go` — `search_pdf` tool description and query parameter
- Search backend: `internal/store/pages.go` — SQLite FTS5 `MATCH` query
- Current description gives no guidance on query format, so the LLM often sends natural language phrases that don't match well with FTS5

## Plan

Update the `search_pdf` tool description and query parameter description in `internal/chat/tools.go` to guide the LLM toward sparse keyword queries.

## Checklist

- [x] Update `search_pdf` description and query parameter in `internal/chat/tools.go`

## Notes

No test needed — this is a string-only change to tool metadata.
