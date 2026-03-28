<script lang="ts">
	import PaperList from '$lib/PaperList.svelte';
	import PdfViewer from '$lib/PdfViewer.svelte';
	import UploadZone from '$lib/UploadZone.svelte';
	import ChatPanel from '$lib/ChatPanel.svelte';
	import { papersStore } from '$lib/papers.svelte';
	import { untrack } from 'svelte';
	import {
		getActivePanel,
		getIsMobile,
		setIsMobile,
		toggleSidebar,
		toggleChat,
		closePanel
	} from '$lib/mobile-layout.svelte';

	let fileInput = $state<HTMLInputElement | null>(null);
	let uploading = $state(false);
	let uploadError = $state<string | null>(null);

	$effect(() => {
		untrack(() => {
			papersStore.load().catch((e) => console.error('Failed to load papers:', e));
		});
	});

	// Track viewport width via matchMedia
	$effect(() => {
		const mql = window.matchMedia('(max-width: 1023px)');
		function onChange(e: MediaQueryList | MediaQueryListEvent) {
			setIsMobile(e.matches);
		}
		onChange(mql);
		mql.addEventListener('change', onChange);
		return () => mql.removeEventListener('change', onChange);
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
		{#if getIsMobile()}
			<button
				class="mobile-toggle sidebar-toggle"
				onclick={toggleSidebar}
				aria-label="Toggle sidebar"
			>&#9776;</button>
		{/if}
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
			{#if getIsMobile() && papersStore.selectedPaper}
				<button
					class="mobile-toggle chat-toggle"
					onclick={toggleChat}
					aria-label="Toggle chat"
				>&#x1F4AC;</button>
			{/if}
		</div>
	</header>
	<div class="app-layout">
		{#if getIsMobile()}
			{#if getActivePanel()}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div class="backdrop" onclick={closePanel}></div>
			{/if}
			<aside
				class="sidebar mobile-overlay from-left"
				class:open={getActivePanel() === 'sidebar'}
			>
				<div class="sidebar-header">Papers</div>
				<PaperList />
				<UploadZone />
			</aside>
		{:else}
			<aside class="sidebar">
				<div class="sidebar-header">Papers</div>
				<PaperList />
				<UploadZone />
			</aside>
		{/if}
		<main class="content">
			{#if papersStore.selectedPaper}
				<PdfViewer paperId={papersStore.selectedPaper.id} />
			{:else}
				<p class="placeholder">Select a paper to view</p>
			{/if}
		</main>
		{#if papersStore.selectedPaper}
			{#if getIsMobile()}
				<div
					class="chat-overlay-wrapper mobile-overlay from-right"
					class:open={getActivePanel() === 'chat'}
				>
					<ChatPanel paperId={papersStore.selectedPaper.id} />
				</div>
			{:else}
				<ChatPanel paperId={papersStore.selectedPaper.id} />
			{/if}
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
		position: relative;
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

	/* Mobile toggle buttons */
	.mobile-toggle {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		border: none;
		background: none;
		color: #fff;
		font-size: 1.25rem;
		cursor: pointer;
		flex-shrink: 0;
		-webkit-tap-highlight-color: transparent;
	}

	.mobile-toggle:hover {
		background: rgba(255, 255, 255, 0.1);
		border-radius: 6px;
	}

	/* Backdrop overlay */
	.backdrop {
		position: fixed;
		inset: 0;
		top: 48px;
		background: rgba(0, 0, 0, 0.4);
		z-index: 90;
	}

	/* Slide-over panel base */
	.mobile-overlay {
		position: fixed;
		top: 48px;
		bottom: 0;
		z-index: 100;
		transition: transform 250ms ease;
	}

	.mobile-overlay.from-left {
		left: 0;
		transform: translateX(-100%);
	}

	.mobile-overlay.from-right {
		right: 0;
		transform: translateX(100%);
	}

	.mobile-overlay.open {
		transform: translateX(0);
	}

	/* Mobile sidebar overlay */
	.sidebar.mobile-overlay {
		width: 280px;
		min-width: 0;
		box-shadow: 2px 0 12px rgba(0, 0, 0, 0.15);
	}

	/* Mobile chat overlay */
	.chat-overlay-wrapper {
		width: 70vw;
		max-width: 400px;
		min-width: 280px;
		display: flex;
		flex-direction: column;
		background: #fff;
		box-shadow: -2px 0 12px rgba(0, 0, 0, 0.15);
	}

	/* On mobile, override chat panel to fill the wrapper */
	.chat-overlay-wrapper :global(.chat-panel) {
		width: 100%;
		min-width: 0;
		border-left: none;
	}

	/* Mobile touch target sizing */
	@media (max-width: 1023px) {
		.upload-btn {
			min-height: 44px;
			min-width: 44px;
			padding: 0.5rem 1rem;
		}

		.content {
			min-width: 0;
		}
	}
</style>
