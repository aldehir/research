# Task 25: Clean up frontend visual/UI tests

Remove frontend tests that verify visual/UI concerns (DOM structure, markup, CSS classes, styling, layout) and keep only tests for business logic (stores, utils, data transforms, API clients).

## Context

18 test files in `frontend/tests/`. Most are pure business logic and need no changes. Four files need cleanup:

- **`chat-panel-structure.test.ts`** — entirely UI/visual (DOM structure, CSS checks). Delete whole file.
- **`pdf-render.test.ts`** — mostly UI/visual (canvas creation, text layer DOM, CSS variables, annotation layer). Keep: `getPageDimensions` tests (dimension calculation), `creates viewport with scale * PDF_TO_CSS_UNITS`, `stops rendering when abort signal fires`. Remove all other tests.
- **`markdown.test.ts`** — "basic formatting", "code blocks", "lists", "other elements" describe blocks are UI/visual (HTML tag rendering). Keep: "XSS sanitization" and "streaming resilience" blocks (security and error handling).
- **`theme.test.ts`** — 2 tests check DOM side effects (`sets data-theme attribute on html element`, `removes data-theme attribute when set to system`). Remove those 2 tests, keep everything else.

No changes needed to the 14 pure business logic test files.

## Checklist

- [x] Delete `chat-panel-structure.test.ts` entirely
- [x] Clean `pdf-render.test.ts`: remove UI/DOM tests, keep `getPageDimensions`, viewport creation, abort signal tests
- [x] Clean `markdown.test.ts`: remove formatting/rendering describe blocks, keep XSS sanitization and streaming resilience
- [x] Clean `theme.test.ts`: remove 2 DOM attribute tests
- [x] Run `pnpm test` to verify remaining tests pass

## Notes

- `pdf-render.test.ts` will need careful surgery — the business logic tests are interspersed with UI tests in the same describe blocks.
- After cleanup, dead test helpers/imports should be removed from modified files.
