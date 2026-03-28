<script lang="ts">
	import { renderMarkdown } from '$lib/markdown';

	interface Props {
		content: string;
	}

	let { content }: Props = $props();
	let html = $derived(renderMarkdown(content));
</script>

<span class="markdown-content">{@html html}</span>

<style>
	.markdown-content {
		white-space: normal;
	}

	/* Headings */
	.markdown-content :global(h1),
	.markdown-content :global(h2),
	.markdown-content :global(h3),
	.markdown-content :global(h4),
	.markdown-content :global(h5),
	.markdown-content :global(h6) {
		margin: 0.6em 0 0.3em;
		line-height: 1.3;
	}

	.markdown-content :global(h1) { font-size: 1.3em; }
	.markdown-content :global(h2) { font-size: 1.15em; }
	.markdown-content :global(h3) { font-size: 1.05em; }

	/* Paragraphs */
	.markdown-content :global(p) {
		margin: 0.4em 0;
	}

	.markdown-content :global(p:first-child) {
		margin-top: 0;
	}

	.markdown-content :global(p:last-child) {
		margin-bottom: 0;
	}

	/* Inline code */
	.markdown-content :global(code) {
		background: rgba(0, 0, 0, 0.06);
		padding: 0.15em 0.35em;
		border-radius: 3px;
		font-size: 0.88em;
		font-family: 'SF Mono', 'Fira Code', 'Fira Mono', Menlo, Consolas, monospace;
	}

	/* Code blocks */
	.markdown-content :global(pre) {
		background: #1e1e2e;
		color: #cdd6f4;
		padding: 0.75em 1em;
		border-radius: 6px;
		overflow-x: auto;
		margin: 0.5em 0;
		font-size: 0.85em;
		line-height: 1.5;
	}

	.markdown-content :global(pre code) {
		background: none;
		padding: 0;
		border-radius: 0;
		font-size: inherit;
		color: inherit;
	}

	/* Blockquotes */
	.markdown-content :global(blockquote) {
		border-left: 3px solid #c0c0c0;
		margin: 0.5em 0;
		padding: 0.25em 0.75em;
		color: #555;
	}

	.markdown-content :global(blockquote p) {
		margin: 0.2em 0;
	}

	/* Lists */
	.markdown-content :global(ul),
	.markdown-content :global(ol) {
		margin: 0.4em 0;
		padding-left: 1.5em;
	}

	.markdown-content :global(li) {
		margin: 0.15em 0;
	}

	.markdown-content :global(li > ul),
	.markdown-content :global(li > ol) {
		margin: 0.1em 0;
	}

	/* Tables */
	.markdown-content :global(table) {
		border-collapse: collapse;
		margin: 0.5em 0;
		font-size: 0.88em;
		width: 100%;
	}

	.markdown-content :global(th),
	.markdown-content :global(td) {
		border: 1px solid #ddd;
		padding: 0.35em 0.6em;
		text-align: left;
	}

	.markdown-content :global(th) {
		background: rgba(0, 0, 0, 0.04);
		font-weight: 600;
	}

	/* Horizontal rules */
	.markdown-content :global(hr) {
		border: none;
		border-top: 1px solid #ddd;
		margin: 0.75em 0;
	}

	/* Links */
	.markdown-content :global(a) {
		color: #1a73e8;
		text-decoration: none;
	}

	.markdown-content :global(a:hover) {
		text-decoration: underline;
	}

	/* Images */
	.markdown-content :global(img) {
		max-width: 100%;
		height: auto;
	}

	/* Strong/emphasis */
	.markdown-content :global(strong) {
		font-weight: 600;
	}
</style>
