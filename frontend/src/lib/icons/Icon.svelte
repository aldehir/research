<script lang="ts">
	interface Props {
		size?: number;
		d: string;
		strokeWidth?: number;
	}

	let { size = 24, d, strokeWidth = 2 }: Props = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width={strokeWidth}
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>{@html d}</svg>
