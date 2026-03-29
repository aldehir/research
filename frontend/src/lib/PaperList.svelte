<script lang="ts">
	import { papersStore } from '$lib/papers.svelte';
	import type { Paper } from '$lib/api';
	import { Icon, X } from '$lib/icons';
	import { goto } from '$app/navigation';

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
		goto(`/papers/${paper.id}`);
	}

	async function handleDelete(event: Event, paper: Paper) {
		event.stopPropagation();
		if (window.confirm(`Delete "${paper.title}"?`)) {
			const wasSelected = papersStore.selectedId === paper.id;
			await papersStore.remove(paper.id);
			if (wasSelected) {
				goto('/');
			}
		}
	}
</script>

<div class="paper-list">
	{#if papersStore.papers.length === 0}
		<p class="empty">No papers uploaded</p>
	{:else}
		<ul>
			{#each papersStore.papers as paper (paper.id)}
				<li>
					<button
						class="paper-item"
						class:selected={papersStore.selectedPaper?.id === paper.id}
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
						<Icon d={X} size={16} />
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
		color: var(--color-text-tertiary);
		text-align: center;
		padding: 2rem 1rem;
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	li {
		position: relative;
		border-bottom: 1px solid var(--color-border);
	}

	.paper-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding: 0.75rem 2.5rem 0.75rem 1rem;
		border: none;
		background: none;
		cursor: pointer;
		text-align: left;
		width: 100%;
		color: var(--color-text);
	}

	.paper-item:hover {
		background: var(--color-surface-hover);
	}

	.paper-item.selected {
		background: var(--color-surface-active);
	}

	.paper-title {
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.paper-meta {
		font-size: 0.8rem;
		color: var(--color-text-secondary);
	}

	.delete-btn {
		position: absolute;
		right: 0;
		top: 0;
		bottom: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0 0.5rem;
		border: none;
		background: none;
		cursor: pointer;
		color: var(--color-text-tertiary);
		line-height: 1;
		z-index: 1;
	}

	.delete-btn:hover {
		color: var(--color-danger);
	}

	@media (max-width: 1023px) {
		.paper-item {
			min-height: 44px;
			padding: 0.75rem 3rem 0.75rem 1rem;
		}

		.delete-btn {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
