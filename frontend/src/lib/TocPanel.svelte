<script lang="ts">
	import type { TocEntry } from '$lib/pdf-outline';
	import { findActiveTocEntry } from '$lib/pdf-outline';

	interface Props {
		entries: TocEntry[];
		currentPage: number;
		onNavigate: (pageNumber: number) => void;
	}

	let { entries, currentPage, onNavigate }: Props = $props();

	let activeEntry = $derived(findActiveTocEntry(entries, currentPage));

	let collapsed = $state(new Set<TocEntry>());

	function toggleCollapse(entry: TocEntry): void {
		const next = new Set(collapsed);
		if (next.has(entry)) {
			next.delete(entry);
		} else {
			next.add(entry);
		}
		collapsed = next;
	}
</script>

{#snippet tocTree(items: TocEntry[], depth: number)}
	{#each items as entry (entry.title + entry.pageNumber)}
		<div
			class="toc-entry"
			class:active={entry === activeEntry}
			data-depth={depth}
			style="padding-left: {0.75 + depth * 0.75}rem"
		>
			<div class="toc-row">
				{#if entry.children.length > 0}
					<button
						class="toc-toggle"
						onclick={() => toggleCollapse(entry)}
						aria-label={collapsed.has(entry) ? 'Expand' : 'Collapse'}
					>
						{collapsed.has(entry) ? '\u25B6' : '\u25BC'}
					</button>
				{:else}
					<span class="toc-toggle-spacer"></span>
				{/if}
				<button class="toc-title" onclick={() => onNavigate(entry.pageNumber)}>
					<span class="toc-title-text">{entry.title}</span>
					<span class="toc-page">{entry.pageNumber}</span>
				</button>
			</div>
		</div>
		{#if entry.children.length > 0 && !collapsed.has(entry)}
			{@render tocTree(entry.children, depth + 1)}
		{/if}
	{/each}
{/snippet}

<div class="toc-panel">
	<div class="toc-header">
		<h3>Contents</h3>
	</div>
	<div class="toc-list">
		{#if entries.length === 0}
			<p class="toc-empty">No table of contents</p>
		{:else}
			{@render tocTree(entries, 0)}
		{/if}
	</div>
</div>

<style>
	.toc-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow: hidden;
		background: #2a2a3e;
		color: #d4d4e8;
	}

	.toc-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #3a3a52;
		flex-shrink: 0;
	}

	.toc-header h3 {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: #e8e8f0;
	}

	.toc-list {
		flex: 1;
		overflow-y: auto;
		padding: 0.25rem 0;
	}

	.toc-entry {
		padding-right: 0.5rem;
	}

	.toc-entry.active {
		background: rgba(99, 132, 255, 0.15);
	}

	.toc-row {
		display: flex;
		align-items: flex-start;
		gap: 0.25rem;
	}

	.toc-toggle {
		flex-shrink: 0;
		border: none;
		background: none;
		color: #888;
		cursor: pointer;
		padding: 0.3rem 0.15rem;
		font-size: 0.55rem;
		line-height: 1;
	}

	.toc-toggle:hover {
		color: #bbb;
	}

	.toc-toggle-spacer {
		display: inline-block;
		width: 0.85rem;
		flex-shrink: 0;
	}

	.toc-title {
		flex: 1;
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		border: none;
		background: none;
		color: inherit;
		cursor: pointer;
		padding: 0.25rem 0.25rem;
		text-align: left;
		font-size: 0.82rem;
		line-height: 1.35;
		border-radius: 3px;
		min-width: 0;
	}

	.toc-title:hover {
		background: rgba(255, 255, 255, 0.08);
		color: #fff;
	}

	.toc-title-text {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.toc-page {
		flex-shrink: 0;
		font-size: 0.75rem;
		color: #777;
	}

	.toc-empty {
		color: #666;
		font-size: 0.85rem;
		text-align: center;
		padding: 2rem 1rem;
		margin: 0;
	}
</style>
