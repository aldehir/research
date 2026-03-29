<script lang="ts">
	import { papersStore } from '$lib/papers.svelte';

	let dragOver = $state(false);
	let uploading = $state(false);
	let error = $state<string | null>(null);
	let fileInput = $state<HTMLInputElement | null>(null);

	async function handleFile(file: File) {
		if (!file.name.toLowerCase().endsWith('.pdf')) {
			error = 'Only PDF files are accepted';
			return;
		}
		error = null;
		uploading = true;
		try {
			await papersStore.upload(file);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Upload failed';
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

	function handleDragLeave() {
		dragOver = false;
	}

	function handleFileChange(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) handleFile(file);
		input.value = '';
	}

	function openFilePicker() {
		fileInput?.click();
	}
</script>

<div
	class="upload-zone"
	class:drag-over={dragOver}
	class:uploading
	ondrop={handleDrop}
	ondragover={handleDragOver}
	ondragleave={handleDragLeave}
	role="button"
	tabindex="0"
	onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') openFilePicker(); }}
	onclick={openFilePicker}
>
	<input
		bind:this={fileInput}
		type="file"
		accept=".pdf"
		onchange={handleFileChange}
		hidden
	/>
	{#if uploading}
		<p>Uploading...</p>
	{:else}
		<p>Drop PDF here or click to upload</p>
	{/if}
	{#if error}
		<p class="error">{error}</p>
	{/if}
</div>

<style>
	.upload-zone {
		border: 2px dashed var(--color-border-strong);
		border-radius: var(--radius);
		padding: 1.5rem 1rem;
		text-align: center;
		cursor: pointer;
		transition: border-color 0.2s, background 0.2s;
		margin: 0.75rem;
	}

	.upload-zone:hover,
	.upload-zone.drag-over {
		border-color: var(--color-primary);
		background: var(--color-primary-light);
	}

	.upload-zone.uploading {
		opacity: 0.7;
		pointer-events: none;
	}

	.upload-zone p {
		margin: 0;
		color: var(--color-text-secondary);
	}

	.error {
		color: var(--color-danger) !important;
		margin-top: 0.5rem !important;
		font-size: 0.85rem;
	}
</style>
