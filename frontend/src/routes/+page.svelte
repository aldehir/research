<script lang="ts">
	import { papersStore } from '$lib/papers.svelte';
	import { onMount } from 'svelte';

	onMount(() => {
		papersStore.deselect();
	});
</script>

<p class="placeholder">Select a paper to view</p>

<style>
	.placeholder {
		color: var(--color-text-tertiary);
		font-size: 1.1rem;
	}
</style>
