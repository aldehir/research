<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';
	import { getSelectedText, clearSelectedText, getCurrentPage } from '$lib/pdf-context.svelte';
	import { Icon, Quote, X, Send } from '$lib/icons';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();
	let inputText = $state('');
	let attachedSelection = $state('');

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			handleSend();
		}
	}

	function captureSelection() {
		const sel = getSelectedText();
		if (sel) {
			attachedSelection = sel;
		}
	}

	function removeSelection() {
		attachedSelection = '';
	}

	async function handleSend() {
		const content = inputText.trim();
		const chatId = getActiveSessionId();
		if (!content || !chatId || getIsStreaming()) return;

		const selectedText = attachedSelection || undefined;
		const currentPage = getCurrentPage();

		inputText = '';
		attachedSelection = '';
		clearSelectedText();

		await sendChatMessage(paperId, chatId, content,
			{ selectedText, currentPage }
		);
	}
</script>

<div class="input-area">
	{#if attachedSelection}
		<div class="selection-chip">
			<span class="chip-text">{attachedSelection}</span>
			<button class="chip-remove" onclick={removeSelection} aria-label="Remove selection"><Icon d={X} size={14} /></button>
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
		<div class="btn-group">
			{#if getSelectedText() && !attachedSelection}
				<button
					class="quote-btn"
					onclick={captureSelection}
					title="Attach selected text"
					aria-label="Attach selected text"
				><Icon d={Quote} size={16} /></button>
			{/if}
			<button
				class="send-btn"
				onclick={handleSend}
				disabled={getIsStreaming() || !inputText.trim() || !getActiveSessionId()}
			>
				<Icon d={Send} size={16} />
			</button>
		</div>
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
		align-items: flex-end;
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

	.btn-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		align-items: stretch;
	}

	.send-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		height: var(--btn-height-lg);
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

	.quote-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		height: var(--btn-height-md);
		padding: 0 0.5rem;
		border: 1px solid var(--color-primary);
		background: var(--color-primary-light);
		color: var(--color-primary);
		border-radius: var(--radius);
		cursor: pointer;
	}

	.quote-btn:hover {
		background: var(--color-surface-active);
	}

	.selection-chip {
		display: flex;
		align-items: flex-start;
		gap: 0.25rem;
		padding: 0.4rem 0.5rem;
		background: var(--color-primary-light);
		border: 1px solid var(--color-primary);
		border-radius: var(--radius);
		margin-bottom: 0.5rem;
	}

	.chip-text {
		flex: 1;
		font-size: 0.8rem;
		color: var(--color-text);
		line-height: 1.3;
		max-height: 3.9rem;
		overflow: hidden;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.chip-remove {
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		background: none;
		cursor: pointer;
		color: var(--color-text-secondary);
		padding: 0;
		flex-shrink: 0;
	}

	.chip-remove:hover {
		color: var(--color-text);
	}

	@media (max-width: 1023px) {
		.send-btn {
			min-width: 44px;
			min-height: 44px;
		}

		.quote-btn {
			min-width: 44px;
			min-height: 44px;
		}

		.chip-remove {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
