<script lang="ts">
	import 'pdfjs-dist/web/pdf_viewer.css';
	import * as pdfjsLib from 'pdfjs-dist';
	import type { PDFDocumentProxy } from 'pdfjs-dist';
	import { getPdfUrl } from '$lib/api';
	import { clampPage, zoomIn, zoomOut, formatZoom, DEFAULT_SCALE } from '$lib/pdf-utils';
	import { extractPageText, extractSurroundingContext } from '$lib/text-context';
	import { setSelection, clearSelection, getSelectedText } from '$lib/selection.svelte';
	import { renderPage } from '$lib/pdf-render';

	pdfjsLib.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.mjs';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();

	let currentPage = $state(1);
	let totalPages = $state(0);
	let scale = $state(DEFAULT_SCALE);
	let pdfDoc: PDFDocumentProxy | null = $state.raw(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let scrollContainer: HTMLDivElement | undefined = $state();
	let pageElements = new Map<number, HTMLDivElement>();
	let renderGeneration = 0;

	let zoomDisplay = $derived(formatZoom(scale));
	let jumpPageInput = $state('');

	// Track what we've loaded to avoid duplicate loads
	let loadedPaperId: string | null = null;
	let currentLoadId = 0;

	async function loadPdf(id: string): Promise<void> {
		// Guard: don't reload the same paper
		if (id === loadedPaperId && pdfDoc) return;

		const thisLoad = ++currentLoadId;
		loadedPaperId = id;
		loading = true;
		error = null;
		pageElements = new Map();
		renderGeneration++;

		if (pdfDoc) {
			pdfDoc.destroy();
			pdfDoc = null;
		}

		try {
			const url = getPdfUrl(id);
			console.log('[PdfViewer] loading', url);
			const doc = await pdfjsLib.getDocument(url).promise;

			// Stale check: another load started while we were waiting
			if (thisLoad !== currentLoadId) {
				doc.destroy();
				return;
			}

			console.log('[PdfViewer] loaded, pages:', doc.numPages);
			pdfDoc = doc;
			totalPages = doc.numPages;
			currentPage = 1;
		} catch (e) {
			if (thisLoad !== currentLoadId) return;
			error = e instanceof Error ? e.message : 'Failed to load PDF';
		} finally {
			if (thisLoad === currentLoadId) {
				loading = false;
			}
		}
	}

	async function renderAllPages(doc: PDFDocumentProxy, currentScale: number): Promise<void> {
		const gen = ++renderGeneration;
		for (let i = 1; i <= doc.numPages; i++) {
			if (renderGeneration !== gen) return;
			const container = pageElements.get(i);
			if (!container) continue;
			const page = await doc.getPage(i);
			if (renderGeneration !== gen) return;
			await renderPage(page, container, currentScale);
		}
	}

	function pageAction(node: HTMLDivElement, pageNum: number) {
		pageElements.set(pageNum, node);

		// When the last page element is registered, trigger render
		if (pdfDoc && pageElements.size === totalPages) {
			renderAllPages(pdfDoc, scale);
		}

		return {
			destroy() {
				pageElements.delete(pageNum);
			}
		};
	}

	function handleScroll(): void {
		if (!scrollContainer) return;

		const containerRect = scrollContainer.getBoundingClientRect();
		const containerCenter = containerRect.top + containerRect.height / 2;
		let closestPage = 1;
		let closestDistance = Infinity;

		for (const [pageNum, el] of pageElements) {
			const rect = el.getBoundingClientRect();
			const pageCenter = rect.top + rect.height / 2;
			const distance = Math.abs(pageCenter - containerCenter);
			if (distance < closestDistance) {
				closestDistance = distance;
				closestPage = pageNum;
			}
		}

		currentPage = closestPage;
	}

	function goToPage(pageNum: number): void {
		const clamped = clampPage(pageNum, totalPages);
		const el = pageElements.get(clamped);
		if (el && scrollContainer) {
			el.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
		currentPage = clamped;
	}

	function handleJumpPage(): void {
		const num = parseInt(jumpPageInput, 10);
		if (!isNaN(num)) {
			goToPage(num);
		}
		jumpPageInput = '';
	}

	function handleZoomIn(): void {
		scale = zoomIn(scale);
		if (pdfDoc) renderAllPages(pdfDoc, scale);
	}

	function handleZoomOut(): void {
		scale = zoomOut(scale);
		if (pdfDoc) renderAllPages(pdfDoc, scale);
	}

	function handleMouseUp(): void {
		const selection = window.getSelection();
		if (!selection || selection.isCollapsed || !selection.toString().trim()) {
			clearSelection();
			return;
		}

		const selected = selection.toString();
		const anchorNode = selection.anchorNode;
		if (!anchorNode) {
			clearSelection();
			return;
		}

		let pageWrapper: HTMLDivElement | null = null;
		for (const [, el] of pageElements) {
			if (el.contains(anchorNode)) {
				pageWrapper = el;
				break;
			}
		}

		if (!pageWrapper) {
			clearSelection();
			return;
		}

		const textLayer = pageWrapper.querySelector('.textLayer') as HTMLDivElement | null;
		if (!textLayer) {
			clearSelection();
			return;
		}

		const pageText = extractPageText(textLayer);
		const surrounding = extractSurroundingContext(selected, pageText);
		setSelection(selected, surrounding);
	}

	// Single effect: only tracks paperId, nothing else
	$effect(() => {
		const id = paperId;
		console.log('[PdfViewer] effect fired, paperId:', id);
		loadPdf(id);
	});
</script>

<div class="pdf-viewer">
	<div class="toolbar">
		<div class="toolbar-group">
			<button
				onclick={() => goToPage(currentPage - 1)}
				disabled={currentPage <= 1}
				aria-label="Previous page"
			>&#9664;</button>
			<span class="page-info">{currentPage} / {totalPages}</span>
			<button
				onclick={() => goToPage(currentPage + 1)}
				disabled={currentPage >= totalPages}
				aria-label="Next page"
			>&#9654;</button>
			<input
				type="text"
				class="page-jump"
				placeholder="Go to"
				bind:value={jumpPageInput}
				onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter') handleJumpPage(); }}
				aria-label="Jump to page"
			/>
		</div>
		<div class="toolbar-group">
			<button onclick={handleZoomOut} disabled={scale <= 0.25} aria-label="Zoom out">&minus;</button>
			<span class="zoom-info">{zoomDisplay}</span>
			<button onclick={handleZoomIn} disabled={scale >= 5.0} aria-label="Zoom in">+</button>
		</div>
	</div>

	<div class="pages-container" bind:this={scrollContainer} onscroll={handleScroll} onmouseup={handleMouseUp}>
		{#if loading}
			<p class="status">Loading PDF...</p>
		{:else if error}
			<p class="status error">Error: {error}</p>
		{:else if pdfDoc}
			{#each Array.from({ length: totalPages }, (_, i) => i + 1) as pageNum (pageNum)}
				<div
					class="page-wrapper"
					use:pageAction={pageNum}
				>
					<!-- Rendered by pdf.js -->
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.pdf-viewer {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
	}

	.toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 1rem;
		background: #f5f5f5;
		border-bottom: 1px solid #ddd;
		flex-shrink: 0;
		gap: 1rem;
	}

	.toolbar-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.toolbar button {
		padding: 0.25rem 0.75rem;
		border: 1px solid #ccc;
		border-radius: 4px;
		background: white;
		cursor: pointer;
		font-size: 0.9rem;
	}

	.toolbar button:disabled {
		opacity: 0.4;
		cursor: default;
	}

	.toolbar button:hover:not(:disabled) {
		background: #e8e8e8;
	}

	.page-info, .zoom-info {
		font-size: 0.85rem;
		min-width: 4em;
		text-align: center;
	}

	.page-jump {
		width: 4rem;
		padding: 0.2rem 0.4rem;
		border: 1px solid #ccc;
		border-radius: 4px;
		font-size: 0.85rem;
		text-align: center;
	}

	.pages-container {
		flex: 1;
		overflow: auto;
		background: #888;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 8px;
	}

	.page-wrapper {
		background: white;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}

	.status {
		color: #ddd;
		font-size: 1.1rem;
		margin-top: 2rem;
	}

	.status.error {
		color: #ff6b6b;
	}
</style>
