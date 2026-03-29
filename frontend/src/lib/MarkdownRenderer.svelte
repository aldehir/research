<script lang="ts">
	import { renderMarkdown } from '$lib/markdown';

	interface Props {
		content: string;
	}

	let { content }: Props = $props();
	let html = $derived(renderMarkdown(content));
	let container: HTMLElement;

	const COPY_ICON = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>';
	const CHECK_ICON = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>';

	function handleCopy(button: HTMLButtonElement, pre: HTMLPreElement) {
		const code = pre.querySelector('code');
		const text = code ? code.textContent ?? '' : pre.textContent ?? '';
		navigator.clipboard.writeText(text).then(() => {
			button.innerHTML = CHECK_ICON;
			button.classList.add('copied');
			setTimeout(() => {
				button.innerHTML = COPY_ICON;
				button.classList.remove('copied');
			}, 1500);
		});
	}

	$effect(() => {
		// Re-run whenever html changes
		void html;
		if (!container) return;

		// Clean up old buttons
		container.querySelectorAll('.copy-btn').forEach(btn => btn.remove());

		const pres = container.querySelectorAll('pre');
		const cleanups: (() => void)[] = [];

		for (const pre of pres) {
			pre.style.position = 'relative';
			const btn = document.createElement('button');
			btn.className = 'copy-btn';
			btn.type = 'button';
			btn.innerHTML = COPY_ICON;
			btn.title = 'Copy code';

			const handler = () => handleCopy(btn, pre);
			btn.addEventListener('click', handler);
			cleanups.push(() => btn.removeEventListener('click', handler));

			pre.appendChild(btn);
		}

		return () => {
			cleanups.forEach(fn => fn());
		};
	});
</script>

<span class="markdown-content" bind:this={container}>{@html html}</span>

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
		background: var(--color-bg-tertiary);
		padding: 0.15em 0.35em;
		border-radius: var(--radius-sm);
		font-size: 0.88em;
		font-family: 'SF Mono', 'Fira Code', 'Fira Mono', Menlo, Consolas, monospace;
	}

	/* Code blocks */
	.markdown-content :global(pre) {
		background: var(--color-code-bg);
		color: var(--color-code-text);
		padding: 0.75em 1em;
		border-radius: var(--radius);
		overflow-x: auto;
		margin: 0.5em 0;
		font-size: 0.85em;
		line-height: 1.5;
		position: relative;
	}

	.markdown-content :global(pre code) {
		background: none;
		padding: 0;
		border-radius: 0;
		font-size: inherit;
		color: inherit;
	}

	/* Copy button */
	.markdown-content :global(.copy-btn) {
		position: absolute;
		top: 0.5em;
		right: 0.5em;
		background: transparent;
		border: 1px solid var(--color-code-border);
		border-radius: var(--radius-sm);
		color: var(--color-code-text);
		cursor: pointer;
		padding: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.15s, background 0.15s;
		line-height: 1;
	}

	.markdown-content :global(pre:hover .copy-btn) {
		opacity: 0.7;
	}

	.markdown-content :global(.copy-btn:hover) {
		opacity: 1 !important;
		background: var(--color-code-hover);
	}

	.markdown-content :global(.copy-btn.copied) {
		opacity: 1 !important;
		color: var(--color-code-string);
	}

	/* Blockquotes */
	.markdown-content :global(blockquote) {
		border-left: 3px solid var(--color-border-strong);
		margin: 0.5em 0;
		padding: 0.25em 0.75em;
		color: var(--color-text-secondary);
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
		border: 1px solid var(--color-border);
		padding: 0.35em 0.6em;
		text-align: left;
	}

	.markdown-content :global(th) {
		background: var(--color-bg-tertiary);
		font-weight: 600;
	}

	/* Horizontal rules */
	.markdown-content :global(hr) {
		border: none;
		border-top: 1px solid var(--color-border);
		margin: 0.75em 0;
	}

	/* Links */
	.markdown-content :global(a) {
		color: var(--color-primary);
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
