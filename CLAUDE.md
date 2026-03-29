# Research Paper Reader

PDF reader web app with LLM chat. Go backend, SvelteKit 5 SPA frontend.

## Project Structure

```
cmd/research-server/ — Go entrypoint (cobra CLI)
internal/
  api/              — HTTP handlers (net/http ServeMux)
  store/            — SQLite data access
  anthropic/        — Anthropic Messages API client
  pdf/              — PDF storage service
frontend/           — SvelteKit 5 app (adapter-static, SPA mode)
  src/lib/          — Components, stores, API client
  src/routes/       — SvelteKit pages
```

## Backend (Go)

- **Router**: `net/http.ServeMux` with Go 1.22+ method/path patterns (e.g. `GET /api/health`)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **Tests**: `go test ./...` — table-driven tests with `stretchr/testify` assert/require
- **Dependencies**: Prefer stdlib. Only add third-party deps when no stdlib equivalent exists.
- **Run**: `go run ./cmd/research-server`

## Frontend (SvelteKit 5)

- **SPA mode**: `adapter-static` with `fallback: 'index.html'`, SSR disabled
- **PDF rendering**: Mozilla pdf.js (`pdfjs-dist`)
- **Package manager**: pnpm (not npm)
- **Tests**: `vitest` — run with `pnpm test` from `frontend/`
- **Dev**: `pnpm dev` from `frontend/`
- **Runes mode**: Use Svelte 5 runes (`$state`, `$derived`, `$effect`) not legacy `$:` syntax

## Theming & Styling

- **CSS variables**: All colors, radii, and button heights are defined as CSS custom properties in `frontend/src/lib/theme.css`. Never hardcode hex colors in components — use `var(--color-*)`.
- **Light/dark mode**: Light theme on `:root`, dark theme on `[data-theme="dark"]`, system fallback via `@media (prefers-color-scheme: dark)`. Theme state managed by `frontend/src/lib/theme.svelte.ts`.
- **Key tokens**: `--color-bg`, `--color-text`, `--color-primary`, `--color-border`, `--color-danger`, `--color-surface-hover`, `--color-surface-active`, `--radius`, `--btn-height-sm/md/lg`. See `theme.css` for the full set.

## Icons

- **No icon library** — icons are self-contained in `frontend/src/lib/icons/`.
- `Icon.svelte` renders an SVG from a path string. `index.ts` exports named path constants (e.g. `Menu`, `ChevronRight`, `Send`).
- Usage: `import { Icon, Menu } from '$lib/icons'; <Icon d={Menu} size={20} />`
- To add a new icon: add an SVG path constant to `index.ts` (24x24 viewBox, stroke-based). Source paths from [lucide.dev](https://lucide.dev).
- Do NOT add a third-party icon package.

## TDD Workflow

Every feature follows RED → GREEN → REFACTOR:

1. **RED**: Write a failing test first
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Clean up while keeping tests green

Do not skip the failing test step. Do not write production code without a test.

## API Conventions

- All endpoints under `/api`
- JSON request/response bodies, `Content-Type: application/json`
- Streaming responses use SSE (`text/event-stream`)
- UUIDs for all entity IDs
- ISO 8601 timestamps
- Errors return `{"error": "message"}` with appropriate HTTP status

## Code Style

- Go: `gofmt`, no `panic` in library code, explicit error returns
- TypeScript: strict mode, no `any`
- Prefer small functions over comments
- No dead code — delete unused code, don't comment it out

## Logging

All backend code must use structured logging via `log/slog`:

- **Every exported function that does meaningful work** should log its actions — at minimum on entry/completion and on errors
- Use `slog.Logger` as a dependency (accept via constructor or parameter), never call `slog.Default()` in library code
- Use structured key-value pairs, not formatted strings: `logger.Info("paper indexed", "paper_id", id, "pages", n)`
- Log levels: `Info` for operations (HTTP requests, indexing, DB writes), `Warn` for recoverable issues, `Error` for failures, `Debug` for internals (tool args, raw responses)
- Include relevant IDs and durations so logs are useful for debugging
