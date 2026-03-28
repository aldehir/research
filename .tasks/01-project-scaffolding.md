# Task 01: Project Scaffolding

Set up the Go backend and SvelteKit frontend project structures.

## Backend

- [ ] Initialize Go module (already have `go.mod`)
- [ ] Create `cmd/server/main.go` with a basic HTTP server
- [ ] Use `net/http.ServeMux` with Go 1.22+ patterns for a health check endpoint (`GET /api/health`)
- [ ] Write test for the health check handler

## Frontend

- [ ] Scaffold SvelteKit 5 app in `frontend/` using `pnpm create svelte`
- [ ] Configure `adapter-static` with `fallback: 'index.html'` (SPA mode)
- [ ] Disable SSR globally
- [ ] Configure Vite to proxy `/api` requests to the Go backend in dev mode
- [ ] Add vitest and write a trivial passing test
- [ ] Verify `pnpm build` produces a static site

## Integration

- [ ] Backend serves the built frontend static files from `frontend/build/`
- [ ] Verify: `go run ./cmd/server` serves both API and frontend
