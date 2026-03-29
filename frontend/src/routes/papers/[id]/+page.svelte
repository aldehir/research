<script lang="ts">
	import PdfViewer from '$lib/PdfViewer.svelte';
	import { papersStore } from '$lib/papers.svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	$effect(() => {
		const id = page.params.id;
		papersStore.loadAndSelect(id).catch(() => {
			goto('/');
		});
	});
</script>

{#if papersStore.selectedPaper}
	<PdfViewer paperId={papersStore.selectedPaper.id} />
{:else if papersStore.loading}
	<p class="placeholder">Loading...</p>
{:else}
	<p class="placeholder">Paper not found</p>
{/if}

<style>
	.placeholder {
		color: var(--color-text-tertiary);
		font-size: 1.1rem;
		display: flex;
		align-items: center;
		justify-content: center;
		flex: 1;
	}
</style>
