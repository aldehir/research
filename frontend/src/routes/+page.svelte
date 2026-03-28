<script lang="ts">
	import PaperList from '$lib/PaperList.svelte';
	import PdfViewer from '$lib/PdfViewer.svelte';
	import UploadZone from '$lib/UploadZone.svelte';
	import ChatPanel from '$lib/ChatPanel.svelte';
	import { loadPapers, getSelectedPaper } from '$lib/papers.svelte';
	import { onMount } from 'svelte';

	onMount(() => {
		loadPapers();
	});
</script>

<div class="app-layout">
	<aside class="sidebar">
		<h2>Papers</h2>
		<PaperList />
		<UploadZone />
	</aside>
	<main class="content">
		{#if getSelectedPaper()}
			<PdfViewer paperId={getSelectedPaper()!.id} />
		{:else}
			<p class="placeholder">Select a paper to view</p>
		{/if}
	</main>
	{#if getSelectedPaper()}
		<ChatPanel paperId={getSelectedPaper()!.id} />
	{/if}
</div>

<style>
	.app-layout {
		display: flex;
		height: 100vh;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.sidebar {
		width: 320px;
		min-width: 260px;
		border-right: 1px solid #ddd;
		display: flex;
		flex-direction: column;
		background: #fafafa;
	}

	.sidebar h2 {
		margin: 0;
		padding: 1rem;
		font-size: 1.1rem;
		border-bottom: 1px solid #eee;
	}

	.content {
		flex: 1;
		display: flex;
		flex-direction: column;
		color: #666;
		min-width: 0;
		overflow: hidden;
	}

	.content:has(.placeholder) {
		align-items: center;
		justify-content: center;
	}

	.placeholder {
		color: #999;
		font-size: 1.1rem;
	}
</style>
