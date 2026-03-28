<script lang="ts">
	import 'pdfjs-dist/web/pdf_viewer.css';
	import * as pdfjsLib from 'pdfjs-dist';
	import type { PDFDocumentProxy, PDFPageProxy } from 'pdfjs-dist';
	import { getPdfUrl } from '$lib/api';
	import { clampPage, zoomIn, zoomOut, zoomByDelta, formatZoom, fitToWidthScale } from '$lib/pdf-utils';
	import { renderPage, renderAnnotations, clearPage, getPageDimensions, PDF_TO_CSS_UNITS } from '$lib/pdf-render';
	import { computeScrollAnchor, restoreScrollTop } from '$lib/pdf-scroll';
	import { setPages, setCurrentPage, setSelectedText } from '$lib/pdf-context.svelte';
	import { extractOutline, type TocEntry } from '$lib/pdf-outline';
	import TocPanel from '$lib/TocPanel.svelte';
	import { consumeNavigateTarget, getNavigateTarget } from '$lib/pdf-navigate.svelte';

	pdfjsLib.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.mjs';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();

	const CONTAINER_PADDING = 16; // 8px padding on each side of .pages-container

	// Minimal link service for AnnotationLayer — handles internal go-to-page
	// destinations and external URLs opening in a new tab.
	const linkService = {
		getDestinationHash: () => '#',
		getAnchorUrl: () => '#',
		addLinkAttributes(link: HTMLAnchorElement, url: string) {
			link.href = url;
			link.target = '_blank';
			link.rel = 'noopener noreferrer';
		},
		goToDestination: async (dest: string | unknown[]) => {
			if (!pdfDoc) return;
			const resolved = typeof dest === 'string'
				? await pdfDoc.getDestination(dest)
				: dest;
			if (!resolved) return;
			const ref = resolved[0];
			const pageIndex = await pdfDoc.getPageIndex(ref);
			goToPage(pageIndex + 1);
		},
		goToPage: (pageNum: number) => goToPage(pageNum),
		navigateTo: () => {},
		executeNamedAction: () => {},
		executeSetOCGState: () => {}
	};

	let currentPage = $state(1);
	let totalPages = $state(0);
	let scale = $state(1.0);
	let pdfDoc: PDFDocumentProxy | null = $state.raw(null);
	let pages: PDFPageProxy[] = $state.raw([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let scrollContainer: HTMLDivElement | undefined = $state();
	let pageElements = new Map<number, HTMLDivElement>();
	let renderedPages = new Set<number>();
	let pageAbortControllers = new Map<number, AbortController>();
	let observer: IntersectionObserver | null = null;
	let renderGeneration = 0;
	let resizeObserver: ResizeObserver | null = null;
	let isFitToWidth = $state(true);

	let tocEntries = $state<TocEntry[]>([]);
	let tocVisible = $state(false);

	let zoomDisplay = $derived(formatZoom(scale));
	let jumpPageInput = $state('');

	// Track what we've loaded to avoid duplicate loads
	let loadedPaperId: string | null = null;
	let currentLoadId = 0;

	function computeFitScale(): number | null {
		if (!scrollContainer || pages.length === 0) return null;
		const pageWidth = pages[0].getViewport({ scale: 1.0 }).width * PDF_TO_CSS_UNITS;
		return fitToWidthScale(scrollContainer.clientWidth, pageWidth, CONTAINER_PADDING);
	}

	function handleFitToWidth(): void {
		const fit = computeFitScale();
		if (fit !== null) {
			scale = fit;
			isFitToWidth = true;
			rerenderVisible();
		}
	}

	function setupObserver(): void {
		observer?.disconnect();
		observer = new IntersectionObserver(handleIntersection, {
			root: scrollContainer,
			rootMargin: '200% 0px'
		});
		for (const [, el] of pageElements) {
			observer.observe(el);
		}
	}

	function cancelPageRender(pageNum: number): void {
		const ac = pageAbortControllers.get(pageNum);
		if (ac) {
			ac.abort();
			pageAbortControllers.delete(pageNum);
		}
	}

	function handleIntersection(entries: IntersectionObserverEntry[]): void {
		for (const entry of entries) {
			const pageNum = Number(entry.target.getAttribute('data-page'));
			if (!pageNum || !pdfDoc) continue;

			if (entry.isIntersecting && !renderedPages.has(pageNum)) {
				const page = pages[pageNum - 1];
				if (page) {
					cancelPageRender(pageNum);
					renderedPages.add(pageNum);
					const el = entry.target as HTMLDivElement;
					const ac = new AbortController();
					pageAbortControllers.set(pageNum, ac);
					renderPage(page, el, scale, ac.signal).then(() => {
						if (!ac.signal.aborted) {
							renderAnnotations(page, el, scale, linkService);
						}
					});
				}
			} else if (!entry.isIntersecting && renderedPages.has(pageNum)) {
				cancelPageRender(pageNum);
				renderedPages.delete(pageNum);
				clearPage(entry.target as HTMLDivElement);
				// Restore placeholder dimensions after clearing
				const page = pages[pageNum - 1];
				if (page) {
					const dims = getPageDimensions(page, scale);
					const el = entry.target as HTMLDivElement;
					el.style.width = `${dims.width}px`;
					el.style.height = `${dims.height}px`;
				}
			}
		}
	}

	async function loadPdf(id: string): Promise<void> {
		if (id === loadedPaperId && pdfDoc) return;

		const thisLoad = ++currentLoadId;
		loadedPaperId = id;
		loading = true;
		error = null;
		for (const [pageNum] of pageAbortControllers) {
			cancelPageRender(pageNum);
		}
		pageElements = new Map();
		renderedPages = new Set();
		renderGeneration++;
		observer?.disconnect();
		observer = null;

		if (pdfDoc) {
			pdfDoc.destroy();
			pdfDoc = null;
			pages = [];
		}
		tocEntries = [];
		tocVisible = false;

		try {
			const url = getPdfUrl(id);
			const doc = await pdfjsLib.getDocument(url).promise;

			if (thisLoad !== currentLoadId) {
				doc.destroy();
				return;
			}

			// Pre-fetch all page objects (lightweight — no rendering yet)
			const allPages: PDFPageProxy[] = [];
			for (let i = 1; i <= doc.numPages; i++) {
				allPages.push(await doc.getPage(i));
			}

			if (thisLoad !== currentLoadId) {
				doc.destroy();
				return;
			}

			pdfDoc = doc;
			pages = allPages;
			totalPages = doc.numPages;
			currentPage = 1;
			isFitToWidth = true;

			// Extract table of contents
			tocEntries = await extractOutline(doc);
			tocVisible = tocEntries.length > 0;

			// Compute initial fit-to-width scale
			if (scrollContainer && allPages.length > 0) {
				const pageWidth = allPages[0].getViewport({ scale: 1.0 }).width * PDF_TO_CSS_UNITS;
				scale = fitToWidthScale(scrollContainer.clientWidth, pageWidth, CONTAINER_PADDING);
			}
		} catch (e) {
			if (thisLoad !== currentLoadId) return;
			error = e instanceof Error ? e.message : 'Failed to load PDF';
		} finally {
			if (thisLoad === currentLoadId) {
				loading = false;
			}
		}
	}

	function collectPageOffsets(): { pageNum: number; top: number; height: number }[] {
		if (!scrollContainer) return [];
		const containerRect = scrollContainer.getBoundingClientRect();
		const scrollTop = scrollContainer.scrollTop;
		const offsets: { pageNum: number; top: number; height: number }[] = [];
		for (const [pageNum, el] of pageElements) {
			const rect = el.getBoundingClientRect();
			offsets.push({
				pageNum,
				top: rect.top - containerRect.top + scrollTop,
				height: rect.height
			});
		}
		offsets.sort((a, b) => a.pageNum - b.pageNum);
		return offsets;
	}

	async function rerenderVisible(): Promise<void> {
		const gen = ++renderGeneration;

		// Cancel all in-flight renders
		for (const [pageNum] of pageAbortControllers) {
			cancelPageRender(pageNum);
		}
		renderedPages = new Set();

		// Capture scroll anchor before resizing
		const anchor = scrollContainer
			? computeScrollAnchor(scrollContainer.scrollTop, collectPageOffsets())
			: null;

		// Resize all placeholders to new scale, clear rendered content
		for (const [pageNum, el] of pageElements) {
			const page = pages[pageNum - 1];
			if (!page) continue;
			const dims = getPageDimensions(page, scale);
			el.innerHTML = '';
			el.style.width = `${dims.width}px`;
			el.style.height = `${dims.height}px`;
		}

		// Restore scroll position from anchor
		if (anchor && scrollContainer) {
			const newOffsets = collectPageOffsets();
			scrollContainer.scrollTop = restoreScrollTop(anchor, newOffsets);
		}

		if (renderGeneration !== gen) return;

		// Re-setup observer to trigger rendering of now-visible pages
		setupObserver();
	}

	function pageAction(node: HTMLDivElement, pageNum: number) {
		pageElements.set(pageNum, node);
		node.setAttribute('data-page', String(pageNum));

		// Set initial placeholder dimensions
		const page = pages[pageNum - 1];
		if (page) {
			const dims = getPageDimensions(page, scale);
			node.style.width = `${dims.width}px`;
			node.style.height = `${dims.height}px`;
		}

		// When all page elements are registered, start observing
		if (pdfDoc && pageElements.size === totalPages) {
			setupObserver();
		}

		return {
			destroy() {
				pageElements.delete(pageNum);
				renderedPages.delete(pageNum);
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
		isFitToWidth = false;
		rerenderVisible();
	}

	function handleZoomOut(): void {
		scale = zoomOut(scale);
		isFitToWidth = false;
		rerenderVisible();
	}

	function handleWheel(e: WheelEvent): void {
		if (!e.ctrlKey && !e.metaKey) return;
		e.preventDefault();
		const newScale = zoomByDelta(scale, e.deltaY);
		if (newScale !== scale) {
			scale = newScale;
			isFitToWidth = false;
			rerenderVisible();
		}
	}

	function handleKeydown(e: KeyboardEvent): void {
		// Skip when focus is in an input
		if ((e.target as HTMLElement).tagName === 'INPUT') return;

		switch (e.key) {
			case 'PageDown':
				e.preventDefault();
				goToPage(currentPage + 1);
				break;
			case 'PageUp':
				e.preventDefault();
				goToPage(currentPage - 1);
				break;
			case 'Home':
				e.preventDefault();
				goToPage(1);
				break;
			case 'End':
				e.preventDefault();
				goToPage(totalPages);
				break;
			case '+':
			case '=':
				if (e.ctrlKey || e.metaKey) {
					e.preventDefault();
					handleZoomIn();
				}
				break;
			case '-':
				if (e.ctrlKey || e.metaKey) {
					e.preventDefault();
					handleZoomOut();
				}
				break;
			case '0':
				if (e.ctrlKey || e.metaKey) {
					e.preventDefault();
					handleFitToWidth();
				}
				break;
		}
	}

	// Sync pages and currentPage to shared pdf-context store
	$effect(() => {
		setPages(pages);
	});

	$effect(() => {
		setCurrentPage(currentPage);
	});

	// Handle navigation requests from chat tool calls
	$effect(() => {
		const target = getNavigateTarget();
		if (target !== null) {
			consumeNavigateTarget();
			goToPage(target);
		}
	});

	function handleSelectionChange(): void {
		const sel = window.getSelection();
		if (!sel || sel.isCollapsed || !scrollContainer) {
			return;
		}
		// Only capture selection within our PDF text layers
		const anchorNode = sel.anchorNode;
		if (anchorNode && scrollContainer.contains(anchorNode)) {
			const text = sel.toString().trim();
			setSelectedText(text);
		}
	}

	$effect(() => {
		document.addEventListener('selectionchange', handleSelectionChange);
		return () => document.removeEventListener('selectionchange', handleSelectionChange);
	});

	$effect(() => {
		const id = paperId;
		loadPdf(id);
	});

	// Keyboard navigation
	$effect(() => {
		const viewer = scrollContainer?.closest('.pdf-viewer');
		if (!viewer) return;
		const el = viewer as HTMLElement;
		el.addEventListener('keydown', handleKeydown as EventListener);
		return () => el.removeEventListener('keydown', handleKeydown as EventListener);
	});

	// Maintain fit-to-width on container resize
	$effect(() => {
		const container = scrollContainer;
		if (!container) return;

		resizeObserver = new ResizeObserver(() => {
			if (!isFitToWidth || pages.length === 0) return;
			const fit = computeFitScale();
			if (fit !== null && Math.abs(fit - scale) > 0.01) {
				scale = fit;
				rerenderVisible();
			}
		});
		resizeObserver.observe(container);

		return () => {
			resizeObserver?.disconnect();
			resizeObserver = null;
			observer?.disconnect();
		};
	});
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="pdf-viewer" tabindex="-1">
	<div class="toolbar">
		<div class="toolbar-group">
			{#if tocEntries.length > 0}
				<button
					onclick={() => tocVisible = !tocVisible}
					class:active={tocVisible}
					aria-label="Table of contents"
					title="Table of contents"
				>&#9776;</button>
				<span class="toolbar-divider"></span>
			{/if}
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
			<button
				onclick={handleFitToWidth}
				class:active={isFitToWidth}
				aria-label="Fit to width"
				title="Fit to width"
			>&#x2194;</button>
		</div>
	</div>

	<div class="viewer-body">
		{#if tocVisible}
			<TocPanel entries={tocEntries} {currentPage} onNavigate={goToPage} />
		{/if}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="pages-container" bind:this={scrollContainer} onscroll={handleScroll} onwheel={handleWheel}>
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
</div>

<style>
	.pdf-viewer {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
		outline: none;
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

	.toolbar button.active {
		background: #dbeafe;
		border-color: #93b4e8;
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

	.toolbar-divider {
		width: 1px;
		height: 1.2rem;
		background: #ccc;
	}

	.viewer-body {
		flex: 1;
		display: flex;
		min-height: 0;
	}

	.viewer-body :global(.toc-panel) {
		width: 260px;
		min-width: 200px;
		flex-shrink: 0;
		border-right: 1px solid #3a3a52;
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
		flex-shrink: 0;
	}

	.page-wrapper :global(canvas) {
		background-color: white;
		display: block;
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
