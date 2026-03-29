<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';
	import { getCurrentPage } from '$lib/pdf-context.svelte';
	import { getPendingAttachments, removeAttachment, consumeAttachments } from '$lib/attachments.svelte';
	import { Icon, Send, X } from '$lib/icons';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();
	let inputText = $state('');
	let expandedPreview = $state<string | null>(null);

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
		if (!content || !chatId || getIsStreaming()) return;

		const currentPage = getCurrentPage();
		const atts = consumeAttachments();

		inputText = '';

		await sendChatMessage(paperId, chatId, content, currentPage, atts.length > 0 ? atts : undefined);
	}
</script>

<div class="input-area">
	{#if attachments.length > 0}
		<div class="attachment-strip">
			{#each attachments as att (att.id)}
				<div class="attachment-thumb">
					<button class="thumb-preview" onclick={() => expandedPreview = expandedPreview === att.id ? null : att.id}>
						<img src="data:image/png;base64,{att.image_data}" alt="Region from page {att.page}" />
						<span class="thumb-label">p.{att.page}</span>
					</button>
					<button class="thumb-dismiss" onclick={() => removeAttachment(att.id)} aria-label="Remove attachment">
						<Icon d={X} size={12} />
					</button>
					{#if expandedPreview === att.id}
						<div class="attachment-popover">
							<img src="data:image/png;base64,{att.image_data}" alt="Region from page {att.page}" />
							{#if att.text}
								<pre class="attachment-text">{att.text}</pre>
							{/if}
						</div>
					{/if}
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
			rows="2"
		></textarea>
		<button
			class="send-btn"
			onclick={handleSend}
			disabled={getIsStreaming() || !inputText.trim() || !getActiveSessionId()}
		>
			<Icon d={Send} size={16} />
		</button>
	</div>
</div>

<style>
	.input-area {
		border-top: 1px solid var(--color-border);
		padding: 0.75rem;
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

	.attachment-popover {
		position: absolute;
		bottom: 100%;
		left: 0;
		margin-bottom: 0.5rem;
		padding: 0.5rem;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		box-shadow: 0 4px 12px var(--color-shadow);
		z-index: 20;
		max-width: 400px;
		max-height: 300px;
		overflow: auto;
	}

	.attachment-popover img {
		max-width: 100%;
		height: auto;
		border-radius: var(--radius-sm);
	}

	.attachment-text {
		margin: 0.5rem 0 0;
		font-size: 0.8rem;
		line-height: 1.4;
		white-space: pre-wrap;
		word-break: break-word;
		color: var(--color-text);
	}

	@media (max-width: 1023px) {
		.send-btn {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
