<script lang="ts">
	import '$lib/theme.css';
	import type { Snippet } from 'svelte';
	import PaperList from '$lib/PaperList.svelte';
	import ChatPanel from '$lib/ChatPanel.svelte';
	import ResizeHandle from '$lib/ResizeHandle.svelte';
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
	import { getTheme, setTheme, initTheme } from '$lib/theme.svelte';
	import {
		getSidebarWidth,
		getChatWidth,
		initPanelWidths,
		handleSidebarResize,
		handleChatResize
	} from '$lib/panel-widths.svelte';
	import { Icon, Menu, MessageSquare, Sun, Moon, Plus } from '$lib/icons';
	import { onMount } from 'svelte';

	let { children }: { children: Snippet } = $props();

	let layoutEl: HTMLDivElement | undefined = $state();
	let fileInput: HTMLInputElement | undefined = $state();
	let dragOver = $state(false);
	let uploading = $state(false);
	let uploadError = $state<string | null>(null);

	async function handleFile(file: File) {
		if (!file.name.toLowerCase().endsWith('.pdf')) {
			uploadError = 'Only PDF files are accepted';
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
		}
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		const file = event.dataTransfer?.files[0];
		if (file) handleFile(file);
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		dragOver = true;
	}

	function handleDragLeave(event: DragEvent) {
		const sidebar = (event.currentTarget as HTMLElement);
		if (!sidebar.contains(event.relatedTarget as Node)) {
			dragOver = false;
		}
	}

	function handleFileInput(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) handleFile(file);
		input.value = '';
	}

	function openFilePicker() {
		fileInput?.click();
	}

	onMount(() => {
		initTheme();
		initPanelWidths();
	});

	function getLayoutWidth(): number {
		return layoutEl?.clientWidth ?? 1200;
	}

	function onSidebarResize(delta: number) {
		handleSidebarResize(delta, getLayoutWidth());
	}

	function onChatResize(delta: number) {
		handleChatResize(delta, getLayoutWidth());
	}

	function cycleTheme() {
		setTheme(getTheme() === 'light' ? 'dark' : 'light');
	}

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
</script>

<div class="app-shell">
	<header class="app-header">
		{#if getIsMobile()}
			<button
				class="mobile-toggle sidebar-toggle"
				onclick={toggleSidebar}
				aria-label="Toggle sidebar"
			><Icon d={Menu} size={20} /></button>
		{/if}
		<h1 class="app-title">Research Reader</h1>
		<div class="header-actions">
			<button
				class="theme-toggle"
				onclick={cycleTheme}
				aria-label="Toggle theme"
				title="Toggle theme"
			>
				{#if getTheme() === 'light'}
					<Icon d={Sun} size={18} />
				{:else}
					<Icon d={Moon} size={18} />
				{/if}
			</button>
			{#if getIsMobile() && papersStore.selectedPaper}
				<button
					class="mobile-toggle chat-toggle"
					onclick={toggleChat}
					aria-label="Toggle chat"
				><Icon d={MessageSquare} size={20} /></button>
			{/if}
		</div>
	</header>
	<div class="app-layout" bind:this={layoutEl}>
		{#if getIsMobile()}
			{#if getActivePanel()}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div class="backdrop" onclick={closePanel}></div>
			{/if}
			<aside
				class="sidebar mobile-overlay from-left"
				class:open={getActivePanel() === 'sidebar'}
				ondrop={handleDrop}
				ondragover={handleDragOver}
				ondragleave={handleDragLeave}
			>
				<input
					bind:this={fileInput}
					type="file"
					accept=".pdf"
					onchange={handleFileInput}
					hidden
				/>
				<div class="sidebar-header">
					<span>Papers</span>
					<button
						class="upload-btn"
						onclick={openFilePicker}
						aria-label="Upload PDF"
						disabled={uploading}
					><Icon d={Plus} size={16} /></button>
				</div>
				{#if uploadError}
					<div class="upload-error">{uploadError}</div>
				{/if}
				<PaperList />
				{#if dragOver}
					<div class="drop-overlay">
						<div class="drop-overlay-content">Drop PDF here</div>
					</div>
				{/if}
			</aside>
		{:else}
			<aside
				class="sidebar"
				style:width="{getSidebarWidth()}px"
				ondrop={handleDrop}
				ondragover={handleDragOver}
				ondragleave={handleDragLeave}
			>
				<input
					bind:this={fileInput}
					type="file"
					accept=".pdf"
					onchange={handleFileInput}
					hidden
				/>
				<div class="sidebar-header">
					<span>Papers</span>
					<button
						class="upload-btn"
						onclick={openFilePicker}
						aria-label="Upload PDF"
						disabled={uploading}
					><Icon d={Plus} size={16} /></button>
				</div>
				{#if uploadError}
					<div class="upload-error">{uploadError}</div>
				{/if}
				<PaperList />
				{#if dragOver}
					<div class="drop-overlay">
						<div class="drop-overlay-content">Drop PDF here</div>
					</div>
				{/if}
			</aside>
			<ResizeHandle onResize={onSidebarResize} side="left" />
		{/if}
		<main class="content">
			{@render children()}
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
				<ResizeHandle onResize={onChatResize} side="right" />
				<ChatPanel paperId={papersStore.selectedPaper.id} chatWidth={getChatWidth()} />
			{/if}
		{/if}
	</div>
</div>

<style>
	.app-shell {
		display: flex;
		flex-direction: column;
		height: 100vh;
	}

	.app-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 1rem;
		height: 48px;
		background: var(--color-header-bg);
		color: var(--color-header-text);
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
		gap: 0.5rem;
	}

	.theme-toggle {
		display: flex;
		align-items: center;
		justify-content: center;
		width: var(--btn-height-md);
		height: var(--btn-height-md);
		border: none;
		background: none;
		color: var(--color-header-text);
		cursor: pointer;
		border-radius: var(--radius);
		transition: background 0.15s;
	}

	.theme-toggle:hover {
		background: var(--color-header-hover);
	}

	.app-layout {
		display: flex;
		flex: 1;
		min-height: 0;
		position: relative;
	}

	.sidebar-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 0.5rem 0.75rem 1rem;
		font-size: 0.85rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-secondary);
		border-bottom: 1px solid var(--color-border);
	}

	.upload-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: var(--btn-height-sm);
		height: var(--btn-height-sm);
		border: 1px solid var(--color-border);
		background: none;
		color: var(--color-text-secondary);
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}

	.upload-btn:hover {
		background: var(--color-surface-hover);
		color: var(--color-primary);
	}

	.upload-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.upload-error {
		padding: 0.4rem 1rem;
		font-size: 0.8rem;
		color: var(--color-danger);
		background: var(--color-bg-secondary);
		border-bottom: 1px solid var(--color-border);
	}

	.drop-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: color-mix(in srgb, var(--color-primary-light) 90%, transparent);
		border: 2px dashed var(--color-primary);
		border-radius: var(--radius);
		z-index: 10;
		pointer-events: none;
	}

	.drop-overlay-content {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--color-primary);
	}

	.sidebar {
		position: relative;
		min-width: 0;
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		background: var(--color-bg-secondary);
		flex-shrink: 0;
	}

	.content {
		flex: 1;
		display: flex;
		flex-direction: column;
		color: var(--color-text-secondary);
		min-width: 0;
		overflow: hidden;
	}

	.content:has(:global(.placeholder)) {
		align-items: center;
		justify-content: center;
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
		color: var(--color-header-text);
		cursor: pointer;
		flex-shrink: 0;
		-webkit-tap-highlight-color: transparent;
	}

	.mobile-toggle:hover {
		background: var(--color-header-hover);
		border-radius: var(--radius);
	}

	/* Backdrop overlay */
	.backdrop {
		position: fixed;
		inset: 0;
		top: 48px;
		background: var(--color-backdrop);
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
		box-shadow: 2px 0 12px var(--color-shadow);
	}

	/* Mobile chat overlay */
	.chat-overlay-wrapper {
		width: 70vw;
		max-width: 400px;
		min-width: 280px;
		display: flex;
		flex-direction: column;
		background: var(--color-bg);
		box-shadow: -2px 0 12px var(--color-shadow);
	}

	/* On mobile, override chat panel to fill the wrapper */
	.chat-overlay-wrapper :global(.chat-panel) {
		width: 100%;
		min-width: 0;
		border-left: none;
	}

	@media (max-width: 1023px) {
		.content {
			min-width: 0;
		}
	}
</style>
