<script lang="ts">
	import { getPapers, getSelectedPaper, selectPaper, remove } from '$lib/papers.svelte';
	import type { Paper } from '$lib/api';

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	function handleSelect(paper: Paper) {
		selectPaper(paper.id);
	}

	async function handleDelete(event: Event, paper: Paper) {
		event.stopPropagation();
		if (window.confirm(`Delete "${paper.title}"?`)) {
			await remove(paper.id);
		}
	}
</script>

<div class="paper-list">
	{#if getPapers().length === 0}
		<p class="empty">No papers uploaded</p>
	{:else}
		<ul>
			{#each getPapers() as paper (paper.id)}
				<li>
					<button
						class="paper-item"
						class:selected={getSelectedPaper()?.id === paper.id}
						onclick={() => handleSelect(paper)}
					>
						<span class="paper-title">{paper.title}</span>
						<span class="paper-meta">
							{formatDate(paper.created_at)} &middot; {formatSize(paper.file_size)}
						</span>
					</button>
					<button
						class="delete-btn"
						onclick={(e) => handleDelete(e, paper)}
						aria-label="Delete {paper.title}"
					>
						&times;
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.paper-list {
		overflow-y: auto;
		flex: 1;
	}

	.empty {
		color: #888;
		text-align: center;
		padding: 2rem 1rem;
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	li {
		display: flex;
		align-items: center;
		border-bottom: 1px solid #eee;
	}

	.paper-item {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding: 0.75rem 1rem;
		border: none;
		background: none;
		cursor: pointer;
		text-align: left;
		width: 100%;
	}

	.paper-item:hover {
		background: #f5f5f5;
	}

	.paper-item.selected {
		background: #e8f0fe;
	}

	.paper-title {
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.paper-meta {
		font-size: 0.8rem;
		color: #666;
	}

	.delete-btn {
		padding: 0.5rem 0.75rem;
		border: none;
		background: none;
		cursor: pointer;
		color: #999;
		font-size: 1.2rem;
		line-height: 1;
	}

	.delete-btn:hover {
		color: #e00;
	}
</style>
