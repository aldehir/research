# Task 29: Embed frontend build in Go binary

Embed the static SvelteKit build output into the Go binary using `//go:embed`, so the server is a single self-contained executable. Add an option to override the embedded files and serve from a filesystem path instead.

## Context

Summary of relevant existing code:

- **`cmd/server/main.go`** — `serveFrontend()` already uses `os.DirFS("frontend/build")` with `http.FileServerFS()` and includes SPA fallback logic (unmatched routes → `index.html`). It operates on `fs.FS`, which is directly compatible with `embed.FS`.
- **`frontend/build/`** — Static output from `pnpm build` (SvelteKit adapter-static). Contains `index.html`, `_app/` (JS/CSS chunks), and `pdf.worker.min.mjs`.
- **No existing `//go:embed` usage** — clean slate.
- **Config is env-var based** — `ADDR`, `DB_PATH`, `PDF_DIR`, `LOG_LEVEL`, etc. No CLI flag parser.
- **`io/fs` already imported** in main.go.
- **`.gitignore` excludes `frontend/build/`** — this is fine; files only need to exist at `go build` time, not in git.

### Design

1. Create `frontend/embed.go` (package `frontend`) with `//go:embed build` directive. The `frontend/` dir becomes both a Node project and a Go package — Go ignores non-`.go` files.
2. Refactor `serveFrontend()` to accept an `fs.FS` parameter instead of hardcoding `os.DirFS`.
3. In `main()`, choose between `frontend.BuildFS` (embedded) and `os.DirFS(path)` based on a `FRONTEND_DIR` env var.
4. When `FRONTEND_DIR` is set, serve from that path; otherwise serve embedded files.

## Checklist

- [x] Create `frontend/embed.go` with `//go:embed build` directive exporting `BuildFS`
- [x] Refactor `serveFrontend()` to accept an `fs.FS` parameter instead of hardcoding the build dir
- [x] Add `FRONTEND_DIR` env var support — when set, use `os.DirFS(path)` instead of embedded FS
- [x] Add `fs.Sub()` to strip the `build/` prefix from the embedded FS before serving
- [x] Log which mode is active (embedded vs directory override) at startup
- [x] Test: `serveFrontend` serves index.html from a provided `fs.FS`
- [x] Test: SPA fallback still works with the refactored handler

## Notes

- `//go:embed` paths are relative to the source file, and `..` is not allowed — placing `embed.go` in `frontend/` is the cleanest way to reach `frontend/build/`.
- The embedded FS root will be `build/...`, so `fs.Sub(BuildFS, "build")` is needed to strip the prefix.
- Build workflow: `cd frontend && pnpm build` then `go build ./cmd/server`. The build dir doesn't need to be in git.
- A build tag (e.g. `//go:build !dev`) could be added later to skip embedding during development, but is out of scope for this task.
