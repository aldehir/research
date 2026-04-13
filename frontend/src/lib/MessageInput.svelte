<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';
	import { getCurrentPage } from '$lib/pdf-context.svelte';
	import { getPendingAttachments, removeAttachment, consumeAttachments, addAttachment } from '$lib/attachments.svelte';
	import type { PendingAttachment } from '$lib/attachments.svelte';
	import { Icon, Send, X } from '$lib/icons';
	import { resizeImage } from '$lib/resize-image';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();
	let inputText = $state('');
	let modalAttachment = $state<PendingAttachment | null>(null);
	let dragOver = $state(false);

	const attachments = $derived(getPendingAttachments());

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			handleSend();
		}
	}

	async function handleSend() {
		const content = inputText.trim();
		const chatId = getActiveSessionId();
		const atts = consumeAttachments();
		if ((!content && atts.length === 0) || !chatId || getIsStreaming()) return;

		const currentPage = getCurrentPage();

		inputText = '';

		await sendChatMessage(paperId, chatId, content, currentPage, atts.length > 0 ? atts : undefined);
	}

	async function addImageBlob(blob: Blob) {
		const base64 = await resizeImage(blob);
		addAttachment({ image_data: base64, text: '', page: 0 });
	}

	function handlePaste(event: ClipboardEvent) {
		const items = event.clipboardData?.items;
		if (!items) return;
		for (const item of items) {
			if (item.type.startsWith('image/')) {
				event.preventDefault();
				const blob = item.getAsFile();
				if (blob) addImageBlob(blob);
				return;
			}
		}
	}

	function handleDragOver(event: DragEvent) {
		if (event.dataTransfer?.types.includes('Files')) {
			event.preventDefault();
			dragOver = true;
		}
	}

	function handleDragLeave() {
		dragOver = false;
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		const files = event.dataTransfer?.files;
		if (!files) return;
		for (const file of files) {
			if (file.type.startsWith('image/')) {
				addImageBlob(file);
			}
		}
	}

	function closeModal() {
		modalAttachment = null;
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}

	function handleModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}
</script>

<div
	class="input-area"
	class:drag-over={dragOver}
	ondragover={handleDragOver}
	ondragleave={handleDragLeave}
	ondrop={handleDrop}
>
	{#if attachments.length > 0}
		<div class="attachment-strip">
			{#each attachments as att (att.id)}
				<div class="attachment-thumb">
					<button class="thumb-preview" onclick={() => modalAttachment = att}>
						<img src="data:image/png;base64,{att.image_data}" alt={att.page ? `Region from page ${att.page}` : 'Pasted image'} />
						<span class="thumb-label">{att.page ? `p.${att.page}` : 'Pasted'}</span>
					</button>
					<button class="thumb-dismiss" onclick={() => removeAttachment(att.id)} aria-label="Remove attachment">
						<Icon d={X} size={12} />
					</button>
				</div>
			{/each}
		</div>
	{/if}
	<div class="input-row">
		<textarea
			bind:value={inputText}
			placeholder={getActiveSessionId() ? 'Type a message...' : 'Select or create a chat first'}
			disabled={getIsStreaming() || !getActiveSessionId()}
			onkeydown={handleKeydown}
			onpaste={handlePaste}
			rows="2"
		></textarea>
		<button
			class="send-btn"
			onclick={handleSend}
			disabled={getIsStreaming() || !getActiveSessionId() || (!inputText.trim() && attachments.length === 0)}
		>
			<Icon d={Send} size={16} />
		</button>
	</div>
</div>

{#if modalAttachment}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-backdrop" onclick={handleBackdropClick} onkeydown={handleModalKeydown}>
		<div class="modal-content">
			<div class="modal-header">
				<span class="modal-title">{modalAttachment.page ? `Region from page ${modalAttachment.page}` : 'Pasted image'}</span>
				<button class="modal-close" onclick={closeModal} aria-label="Close">
					<Icon d={X} size={18} />
				</button>
			</div>
			<div class="modal-body">
				<img src="data:image/png;base64,{modalAttachment.image_data}" alt={modalAttachment.page ? `Region from page ${modalAttachment.page}` : 'Pasted image'} />
				{#if modalAttachment.text}
					<pre class="modal-text">{modalAttachment.text}</pre>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.input-area {
		border-top: 1px solid var(--color-border);
		padding: 0.75rem;
		transition: background 0.15s;
	}

	.input-area.drag-over {
		background: var(--color-surface-hover);
		outline: 2px dashed var(--color-primary);
		outline-offset: -2px;
	}

	.input-row {
		display: flex;
		gap: 0.5rem;
		align-items: stretch;
	}

	textarea {
		flex: 1;
		resize: none;
		padding: 0.5rem;
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius);
		font-family: inherit;
		font-size: 0.9rem;
		line-height: 1.4;
		background: var(--color-bg);
		color: var(--color-text);
	}

	textarea:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	textarea:disabled {
		background: var(--color-bg-tertiary);
		cursor: not-allowed;
	}

	.send-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0 0.75rem;
		border: none;
		background: var(--color-primary);
		color: var(--color-primary-text);
		border-radius: var(--radius);
		cursor: pointer;
		font-weight: 500;
	}

	.send-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.send-btn:disabled {
		background: var(--color-border-strong);
		cursor: not-allowed;
	}

	.attachment-strip {
		display: flex;
		gap: 0.5rem;
		padding-bottom: 0.5rem;
		overflow-x: auto;
	}

	.attachment-thumb {
		position: relative;
		flex-shrink: 0;
	}

	.thumb-preview {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.2rem;
		padding: 0.25rem;
		background: var(--color-bg-tertiary);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		cursor: pointer;
	}

	.thumb-preview:hover {
		border-color: var(--color-primary);
	}

	.thumb-preview img {
		max-width: 80px;
		max-height: 60px;
		border-radius: 2px;
		object-fit: contain;
	}

	.thumb-label {
		font-size: 0.7rem;
		color: var(--color-text-secondary);
	}

	.thumb-dismiss {
		position: absolute;
		top: -4px;
		right: -4px;
		width: 18px;
		height: 18px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		border: 1px solid var(--color-border);
		background: var(--color-bg);
		color: var(--color-text-secondary);
		cursor: pointer;
		padding: 0;
	}

	.thumb-dismiss:hover {
		background: var(--color-danger);
		color: white;
		border-color: var(--color-danger);
	}

	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: oklch(0 0 0 / 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.modal-content {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		box-shadow: 0 8px 32px var(--color-shadow);
		max-width: min(90vw, 600px);
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.modal-title {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--color-text);
	}

	.modal-close {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.25rem;
		border: none;
		background: none;
		color: var(--color-text-secondary);
		cursor: pointer;
		border-radius: var(--radius-sm);
	}

	.modal-close:hover {
		background: var(--color-surface-hover);
		color: var(--color-text);
	}

	.modal-body {
		padding: 1rem;
		overflow-y: auto;
	}

	.modal-body img {
		max-width: 100%;
		height: auto;
		border-radius: var(--radius-sm);
	}

	.modal-text {
		margin: 0.75rem 0 0;
		padding: 0.75rem;
		font-size: 0.85rem;
		line-height: 1.5;
		white-space: pre;
		overflow-x: auto;
		color: var(--color-text);
		background: var(--color-bg-tertiary);
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-border);
	}

	@media (max-width: 1023px) {
		.send-btn {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
