<script lang="ts">
	import 'pdfjs-dist/web/pdf_viewer.css';
	import * as pdfjsLib from 'pdfjs-dist';
	import type { PDFDocumentProxy, PDFPageProxy } from 'pdfjs-dist';
	import { TextLayer } from 'pdfjs-dist';
	import pdfWorkerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
	import { getPdfUrl } from '$lib/api';
	import { clampPage, zoomIn, zoomOut, formatZoom, DEFAULT_SCALE } from '$lib/pdf-utils';
	import { extractPageText, extractSurroundingContext } from '$lib/text-context';
	import { setSelection, clearSelection, getSelectedText } from '$lib/selection.svelte';

	pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorkerUrl;

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();

	let currentPage = $state(1);
	let totalPages = $state(0);
	let scale = $state(DEFAULT_SCALE);
	let pdfDoc = $state<PDFDocumentProxy | null>(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let scrollContainer: HTMLDivElement | undefined = $state();
	let pageElements = new Map<number, HTMLDivElement>();
	let renderGeneration = 0;

	let zoomDisplay = $derived(formatZoom(scale));
	let jumpPageInput = $state('');
	let hasSelection = $state(false);

	async function loadPdf(id: string): Promise<void> {
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
			const doc = await pdfjsLib.getDocument(url).promise;
			pdfDoc = doc;
			totalPages = doc.numPages;
			currentPage = 1;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load PDF';
		} finally {
			loading = false;
		}
	}

	async function renderPage(
		page: PDFPageProxy,
		container: HTMLDivElement,
		currentScale: number
	): Promise<void> {
		const viewport = page.getViewport({ scale: currentScale });

		container.innerHTML = '';
		container.style.width = `${viewport.width}px`;
		container.style.height = `${viewport.height}px`;
		container.style.position = 'relative';

		const canvas = document.createElement('canvas');
		const dpr = window.devicePixelRatio || 1;
		canvas.width = Math.floor(viewport.width * dpr);
		canvas.height = Math.floor(viewport.height * dpr);
		canvas.style.width = `${viewport.width}px`;
		canvas.style.height = `${viewport.height}px`;

		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		ctx.scale(dpr, dpr);
		container.appendChild(canvas);

		await page.render({ canvasContext: ctx, viewport }).promise;

		const textContent = await page.getTextContent();
		const textDiv = document.createElement('div');
		textDiv.className = 'textLayer';
		container.appendChild(textDiv);

		const textLayer = new TextLayer({
			textContentSource: textContent,
			container: textDiv,
			viewport
		});
		await textLayer.render();
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
	}

	function handleZoomOut(): void {
		scale = zoomOut(scale);
	}

	function handleMouseUp(): void {
		const selection = window.getSelection();
		if (!selection || selection.isCollapsed || !selection.toString().trim()) {
			clearSelection();
			hasSelection = false;
			return;
		}

		const selected = selection.toString();
		const anchorNode = selection.anchorNode;
		if (!anchorNode) {
			clearSelection();
			hasSelection = false;
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
			hasSelection = false;
			return;
		}

		const textLayer = pageWrapper.querySelector('.textLayer') as HTMLDivElement | null;
		if (!textLayer) {
			clearSelection();
			hasSelection = false;
			return;
		}

		const pageText = extractPageText(textLayer);
		const surrounding = extractSurroundingContext(selected, pageText);
		setSelection(selected, surrounding);
		hasSelection = true;
	}

	$effect(() => {
		loadPdf(paperId);
	});

	$effect(() => {
		const doc = pdfDoc;
		const s = scale;
		if (doc && pageElements.size > 0) {
			renderAllPages(doc, s);
		}
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

	{#if hasSelection}
		<div class="selection-indicator">
			Text selected &mdash; use chat to ask about it
		</div>
	{/if}

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

	.page-wrapper :global(.textLayer) {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		overflow: hidden;
		line-height: 1;
	}

	.selection-indicator {
		padding: 0.4rem 1rem;
		background: #e3f2fd;
		color: #1565c0;
		font-size: 0.8rem;
		text-align: center;
		border-bottom: 1px solid #90caf9;
		flex-shrink: 0;
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
