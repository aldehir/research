# Task 20: URL-based paper routing

Update the address bar when viewing a paper and navigate directly to a paper when accessed via URL.

## Context

- **SPA routing**: App uses `adapter-static` with `fallback: 'index.html'` — the Go server already serves `index.html` for any unmatched route (SPA fallback in `cmd/server/main.go:105-123`)
- **Current state**: The app is a single `+page.svelte` route. No URL changes happen when selecting a paper. There's no `history.pushState`/`replaceState` or SvelteKit `goto` usage anywhere in the frontend
- **Paper selection**: `papersStore` (`frontend/src/lib/papers.svelte.ts`) has `selectedId` state and a `select(id)` method. `PaperList.svelte` calls `papersStore.select(paper.id)` on click
- **API**: `getPaper(id)` exists in `$lib/api.ts` to fetch a single paper by UUID
- **Route pattern**: Use `/papers/:id` — matches the API pattern (`/api/papers/{id}`) and is clean for direct linking
- **SvelteKit routing**: Can add a `frontend/src/routes/papers/[id]/+page.svelte` route, or use `history.pushState`/`replaceState` in the SPA page. Since SvelteKit supports filesystem routing with adapter-static, a dedicated route is cleaner
- **Mobile layout**: `+page.svelte` has mobile panel logic (`mobile-layout.svelte`) that must work on both the index and paper routes

## Checklist

- [x] Add SvelteKit route `frontend/src/routes/papers/[id]/+page.svelte` that reads the paper ID from params
- [x] On mount, load the paper by ID (fetch from API if papers list not loaded yet) and set `papersStore.selectedId`
- [x] Update `papersStore.select()` or `PaperList.svelte` to use `goto('/papers/${id}')` so the URL updates on paper selection
- [x] When deselecting / going back to no paper, navigate to `/`
- [x] Handle browser back/forward navigation correctly (popstate)
- [x] Extract shared layout (header, sidebar, mobile panels) so both `/` and `/papers/[id]` share it
- [x] Test: navigating to `/papers/<valid-uuid>` loads and displays that paper
- [x] Test: selecting a paper from the list updates the URL to `/papers/<id>`
- [x] Test: browser back from `/papers/<id>` returns to `/` with no paper selected

## Notes

- The shared layout (header, sidebar, mobile toggles) currently lives in `+page.svelte`. It should move to a `+layout.svelte` so both routes inherit it.
- Need to handle the case where a user navigates directly to `/papers/<id>` — papers list may not be loaded yet, so we need to call `papersStore.load()` and then select.
- Consider whether to use SvelteKit's `goto()` or raw `history.pushState` — `goto()` is idiomatic and handles SvelteKit's internal routing correctly.
