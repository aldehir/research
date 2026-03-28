<script lang="ts">
	import PaperList from '$lib/PaperList.svelte';
	import PdfViewer from '$lib/PdfViewer.svelte';
	import UploadZone from '$lib/UploadZone.svelte';
	import ChatPanel from '$lib/ChatPanel.svelte';
	import { papersStore } from '$lib/papers.svelte';
	import { untrack } from 'svelte';

	let fileInput = $state<HTMLInputElement | null>(null);
	let uploading = $state(false);
	let uploadError = $state<string | null>(null);

	$effect(() => {
		untrack(() => {
			papersStore.load().catch((e) => console.error('Failed to load papers:', e));
		});
	});

	async function handleHeaderUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		if (!file.name.toLowerCase().endsWith('.pdf')) {
			uploadError = 'Only PDF files are accepted';
			input.value = '';
			return;
		}
		uploadError = null;
		uploading = true;
		try {
			await papersStore.upload(file);
		} catch (e) {
			uploadError = e instanceof Error ? e.message : 'Upload failed';
		} finally {
			uploading = false;
			input.value = '';
		}
	}
</script>

<div class="app-shell">
	<header class="app-header">
		<h1 class="app-title">Research Reader</h1>
		<div class="header-actions">
			{#if uploadError}
				<span class="header-error">{uploadError}</span>
			{/if}
			<input
				bind:this={fileInput}
				type="file"
				accept=".pdf"
				onchange={handleHeaderUpload}
				hidden
			/>
			<button
				class="upload-btn"
				onclick={() => fileInput?.click()}
				disabled={uploading}
			>
				{uploading ? 'Uploading...' : 'Upload PDF'}
			</button>
		</div>
	</header>
	<div class="app-layout">
		<aside class="sidebar">
			<div class="sidebar-header">Papers</div>
			<PaperList />
			<UploadZone />
		</aside>
		<main class="content">
			{#if papersStore.selectedPaper}
				<PdfViewer paperId={papersStore.selectedPaper.id} />
			{:else}
				<p class="placeholder">Select a paper to view</p>
			{/if}
		</main>
		{#if papersStore.selectedPaper}
			<ChatPanel paperId={papersStore.selectedPaper.id} />
		{/if}
	</div>
</div>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
	}

	.app-shell {
		display: flex;
		flex-direction: column;
		height: 100vh;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.app-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 1rem;
		height: 48px;
		background: #1a1a2e;
		color: #fff;
		flex-shrink: 0;
	}

	.app-title {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 600;
		letter-spacing: -0.01em;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.header-error {
		color: #ff8a80;
		font-size: 0.8rem;
	}

	.upload-btn {
		padding: 0.35rem 0.9rem;
		border: 1px solid rgba(255, 255, 255, 0.3);
		border-radius: 5px;
		background: transparent;
		color: #fff;
		font-size: 0.85rem;
		cursor: pointer;
		transition: background 0.15s;
	}

	.upload-btn:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.1);
	}

	.upload-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.app-layout {
		display: flex;
		flex: 1;
		min-height: 0;
	}

	.sidebar-header {
		padding: 0.75rem 1rem;
		font-size: 0.85rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #666;
		border-bottom: 1px solid #eee;
	}

	.sidebar {
		width: 280px;
		min-width: 220px;
		border-right: 1px solid #ddd;
		display: flex;
		flex-direction: column;
		background: #fafafa;
	}

	.content {
		flex: 1;
		display: flex;
		flex-direction: column;
		color: #666;
		min-width: 300px;
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
