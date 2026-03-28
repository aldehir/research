<script lang="ts">
	import PaperList from '$lib/PaperList.svelte';
	import UploadZone from '$lib/UploadZone.svelte';
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
			<p>Selected: {getSelectedPaper()?.title}</p>
			<p class="placeholder">PDF viewer coming soon...</p>
		{:else}
			<p class="placeholder">Select a paper to view</p>
		{/if}
	</main>
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
		align-items: center;
		justify-content: center;
		flex-direction: column;
		color: #666;
	}

	.placeholder {
		color: #999;
		font-size: 1.1rem;
	}
</style>
