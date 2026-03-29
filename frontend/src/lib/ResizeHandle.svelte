<script lang="ts">
	interface Props {
		onResize: (delta: number) => void;
		side?: 'left' | 'right';
	}

	let { onResize, side = 'left' }: Props = $props();
	let dragging = $state(false);
	let startX = 0;

	function handlePointerDown(e: PointerEvent) {
		e.preventDefault();
		dragging = true;
		startX = e.clientX;
		const target = e.currentTarget as HTMLElement;
		target.setPointerCapture(e.pointerId);
	}

	function handlePointerMove(e: PointerEvent) {
		if (!dragging) return;
		const delta = e.clientX - startX;
		startX = e.clientX;
		// For panels on the right side, drag direction is inverted
		onResize(side === 'right' ? -delta : delta);
	}

	function handlePointerUp(e: PointerEvent) {
		if (!dragging) return;
		dragging = false;
		const target = e.currentTarget as HTMLElement;
		target.releasePointerCapture(e.pointerId);
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="resize-handle"
	class:dragging
	onpointerdown={handlePointerDown}
	onpointermove={handlePointerMove}
	onpointerup={handlePointerUp}
	onpointercancel={handlePointerUp}
>
	<div class="resize-indicator"></div>
</div>

<style>
	.resize-handle {
		width: 6px;
		cursor: col-resize;
		position: relative;
		flex-shrink: 0;
		z-index: 10;
		touch-action: none;
		user-select: none;
	}

	.resize-indicator {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 2px;
		width: 2px;
		border-radius: 1px;
		background: transparent;
		transition: background 0.15s;
	}

	.resize-handle:hover .resize-indicator,
	.resize-handle.dragging .resize-indicator {
		background: var(--color-primary);
	}

	.resize-handle.dragging {
		cursor: col-resize;
	}

	@media (pointer: coarse) {
		.resize-handle {
			padding-inline: 19px;
			margin-inline: -19px;
		}

		/* Thicker, always-visible bar on touch devices */
		.resize-indicator {
			left: 20px;
			width: 4px;
			border-radius: 2px;
			background: var(--color-border);
		}

		.resize-handle.dragging .resize-indicator {
			background: var(--color-primary);
		}
	}
</style>
