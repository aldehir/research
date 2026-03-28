<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';
	import { getSelectedText, clearSelectedText, getCurrentPage } from '$lib/pdf-context.svelte';

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
			<button class="chip-remove" onclick={removeSelection} aria-label="Remove selection">&times;</button>
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
				>&#x201C;</button>
			{/if}
			<button
				class="send-btn"
				onclick={handleSend}
				disabled={getIsStreaming() || !inputText.trim() || !getActiveSessionId()}
			>
				Send
			</button>
		</div>
	</div>
</div>

<style>
	.input-area {
		border-top: 1px solid #ddd;
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
		border: 1px solid #ccc;
		border-radius: 6px;
		font-family: inherit;
		font-size: 0.9rem;
		line-height: 1.4;
	}

	textarea:focus {
		outline: none;
		border-color: #4285f4;
	}

	textarea:disabled {
		background: #f5f5f5;
		cursor: not-allowed;
	}

	.btn-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		align-items: stretch;
	}

	.send-btn {
		padding: 0.5rem 1rem;
		border: none;
		background: #4285f4;
		color: white;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 500;
	}

	.send-btn:hover:not(:disabled) {
		background: #3367d6;
	}

	.send-btn:disabled {
		background: #ccc;
		cursor: not-allowed;
	}

	.quote-btn {
		padding: 0.25rem 0.75rem;
		border: 1px solid #4285f4;
		background: #e8f0fe;
		color: #4285f4;
		border-radius: 6px;
		cursor: pointer;
		font-size: 1.1rem;
		font-weight: bold;
		line-height: 1;
	}

	.quote-btn:hover {
		background: #d2e3fc;
	}

	.selection-chip {
		display: flex;
		align-items: flex-start;
		gap: 0.25rem;
		padding: 0.4rem 0.5rem;
		background: #e8f0fe;
		border: 1px solid #c5d8f8;
		border-radius: 6px;
		margin-bottom: 0.5rem;
	}

	.chip-text {
		flex: 1;
		font-size: 0.8rem;
		color: #333;
		line-height: 1.3;
		max-height: 3.9rem;
		overflow: hidden;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.chip-remove {
		border: none;
		background: none;
		cursor: pointer;
		font-size: 1rem;
		color: #666;
		padding: 0;
		line-height: 1;
		flex-shrink: 0;
	}

	.chip-remove:hover {
		color: #333;
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
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}
</style>
