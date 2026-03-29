<script lang="ts">
	import { PDF_TO_CSS_UNITS } from '$lib/pdf-render';

	interface Props {
		pagesContainer: HTMLDivElement;
		pageElements: Map<number, HTMLDivElement>;
		scale: number;
		onSelect: (region: { page: number; x: number; y: number; w: number; h: number }) => void;
		onCancel: () => void;
	}

	let { pagesContainer, pageElements, scale, onSelect, onCancel }: Props = $props();

	let selecting = $state(false);
	let startX = 0;
	let startY = 0;
	let currentX = $state(0);
	let currentY = $state(0);
	let rectX = $state(0);
	let rectY = $state(0);
	let rectW = $state(0);
	let rectH = $state(0);

	let overlayEl: HTMLDivElement | undefined = $state();

	// Find which page element contains the given client coordinates
	function findPage(clientX: number, clientY: number): { pageNum: number; el: HTMLDivElement } | null {
		for (const [pageNum, el] of pageElements) {
			const rect = el.getBoundingClientRect();
			if (clientX >= rect.left && clientX <= rect.right &&
				clientY >= rect.top && clientY <= rect.bottom) {
				return { pageNum, el };
			}
		}
		return null;
	}

	// Convert client coords to overlay-local coords (viewport-relative)
	function toOverlay(clientX: number, clientY: number): { x: number; y: number } {
		if (!overlayEl) return { x: 0, y: 0 };
		const rect = overlayEl.getBoundingClientRect();
		return { x: clientX - rect.left, y: clientY - rect.top };
	}

	let activePage: { pageNum: number; el: HTMLDivElement } | null = null;

	function handlePointerDown(e: PointerEvent) {
		const page = findPage(e.clientX, e.clientY);
		if (!page) return;

		e.preventDefault();
		(e.target as HTMLElement).setPointerCapture(e.pointerId);
		activePage = page;
		selecting = true;

		const pos = toOverlay(e.clientX, e.clientY);
		startX = pos.x;
		startY = pos.y;
		currentX = startX;
		currentY = startY;
		updateRect();
	}

	function handlePointerMove(e: PointerEvent) {
		if (!selecting) return;
		e.preventDefault();

		const pos = toOverlay(e.clientX, e.clientY);
		currentX = pos.x;
		currentY = pos.y;
		updateRect();
	}

	function handlePointerUp(e: PointerEvent) {
		if (!selecting || !activePage) return;
		e.preventDefault();
		selecting = false;

		// Minimum size check (ignore tiny accidental drags)
		if (rectW < 5 || rectH < 5) {
			activePage = null;
			return;
		}

		// Convert viewport-relative rect to page-relative PDF points.
		// The page element's getBoundingClientRect gives viewport coords,
		// and the overlay's rect also uses viewport coords, so we can
		// subtract directly.
		const overlayRect = overlayEl!.getBoundingClientRect();
		const pageRect = activePage.el.getBoundingClientRect();

		// Rect position in viewport coords
		const rectViewX = overlayRect.left + rectX;
		const rectViewY = overlayRect.top + rectY;

		// Position relative to the page element
		const relX = rectViewX - pageRect.left;
		const relY = rectViewY - pageRect.top;

		const pdfScale = PDF_TO_CSS_UNITS * scale;
		const pdfX = Math.max(0, Math.round(relX / pdfScale));
		const pdfY = Math.max(0, Math.round(relY / pdfScale));
		const pdfW = Math.round(rectW / pdfScale);
		const pdfH = Math.round(rectH / pdfScale);

		onSelect({
			page: activePage.pageNum,
			x: pdfX,
			y: pdfY,
			w: pdfW,
			h: pdfH
		});

		activePage = null;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			selecting = false;
			activePage = null;
			onCancel();
		}
	}

	function updateRect() {
		rectX = Math.min(startX, currentX);
		rectY = Math.min(startY, currentY);
		rectW = Math.abs(currentX - startX);
		rectH = Math.abs(currentY - startY);
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="region-select-overlay"
	bind:this={overlayEl}
	onpointerdown={handlePointerDown}
	onpointermove={handlePointerMove}
	onpointerup={handlePointerUp}
>
	{#if selecting && rectW > 0 && rectH > 0}
		<div
			class="selection-rect"
			style="left:{rectX}px;top:{rectY}px;width:{rectW}px;height:{rectH}px"
		></div>
	{/if}
</div>

<style>
	.region-select-overlay {
		position: absolute;
		inset: 0;
		cursor: crosshair;
		z-index: 10;
	}

	.selection-rect {
		position: absolute;
		border: 2px dashed var(--color-primary);
		background: oklch(from var(--color-primary) l c h / 0.08);
		pointer-events: none;
	}
</style>
