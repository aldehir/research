# Task 27: Add syntax highlight support + copy button for code blocks

Add syntax highlighting for fenced code blocks in chat responses and a copy-to-clipboard button on each block.

## Context

- Markdown rendering flows through `markdown.ts` (`marked` + `dompurify`) → `MarkdownRenderer.svelte` → `{@html}`
- `marked` v17 is already configured with GFM; code blocks render as plain `<pre><code class="language-*">` with no highlighting
- Code block styling lives in `MarkdownRenderer.svelte` using `:global()` selectors and theme variables (`--color-code-bg`, `--color-code-text`)
- No syntax highlighting library or copy-to-clipboard functionality exists yet
- DOMPurify config will need to allow the CSS classes the highlighter adds

Key files:
- `frontend/src/lib/markdown.ts` — `renderMarkdown()`, marked config
- `frontend/src/lib/MarkdownRenderer.svelte` — renders HTML, owns code block styles
- `frontend/src/lib/theme.css` — CSS custom properties for code colors

## Checklist

- [x] Install highlight.js; configure marked renderer to apply highlighting to fenced code blocks in `markdown.ts`
- [x] Update DOMPurify config to allow highlight.js class attributes on `<span>` elements
- [x] Import a highlight.js theme (dark, matching existing `--color-code-bg`) and verify code blocks render with syntax colors
- [x] Add a copy button overlay to `<pre>` blocks in `MarkdownRenderer.svelte` (position: absolute, top-right corner)
- [x] Implement clipboard copy via `navigator.clipboard.writeText` with visual feedback (icon swap or brief "Copied!" label)
- [x] Ensure styling works in both light and dark themes

## Notes

- highlight.js is the simplest option — auto-detection covers most LLM output languages; shiki is heavier and SSR-oriented
- The copy button needs to be injected post-render since code blocks come from `{@html}`; use an `$effect` to query `<pre>` elements and attach buttons
- Keep DOMPurify strict — only add the minimum allowed attributes/tags needed for highlighting
