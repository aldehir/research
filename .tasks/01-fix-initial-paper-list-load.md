# Task 01: Fix papers not displaying on initial page load

Papers don't appear in the left sidebar when first visiting the app. They only appear after uploading a new paper.

## Context

- **Paper store**: `frontend/src/lib/papers.svelte.ts` — Svelte 5 class-based store using `$state<Paper[]>([])`. The `load()` method fetches papers via `listPapers()` and assigns to `this.papers`.
- **Initial load**: `frontend/src/routes/+page.svelte` lines 13-15 — `onMount` fires `papersStore.load().catch(...)` without awaiting.
- **Post-upload load**: `papersStore.upload()` properly `await`s both `uploadPaper(file)` and `this.load()`, which is why papers appear after upload.
- **PaperList component**: `frontend/src/lib/PaperList.svelte` — reads `papersStore.papers` reactively, shows "No papers uploaded" when empty.
- **API client**: `frontend/src/lib/api.ts` — `listPapers()` fetches `GET /api/papers`.
- **Backend**: `internal/api/papers.go` `handleListPapers` — returns `[]` (not null) for empty result, 200 with JSON array otherwise.
- **Routing**: `internal/api/api.go` registers `GET /api/papers` before the SPA catch-all in `cmd/server/main.go`.

The `onMount` callback fires and forgets `papersStore.load()`. Since `papers` is `$state`, the reactivity should update `PaperList` when the promise resolves. The fact that it doesn't suggests either the initial API call is failing silently, or there's a Svelte 5 reactivity issue with the fire-and-forget async pattern in `onMount`.

## Checklist

- [x] Add test: verify `papersStore.load()` populates papers from API response
- [x] Debug/fix the initial load — likely candidates:
  - Make `onMount` async and await the load
  - Switch to `$effect` for initial data fetch
  - Add loading state to surface silent failures
- [ ] Verify papers appear on initial page visit (manual test)

## Notes

- Check browser console for errors on first load — the `.catch(e => console.error(...))` may be swallowing an error.
- The upload flow works because it properly awaits; the initial load does not.
- SvelteKit SSR is disabled (`+layout.ts` has `ssr = false`), so all rendering is client-side.
