<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';
	import { getSelectedText, getSurroundingText, clearSelection } from '$lib/selection.svelte';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();
	let inputText = $state('');

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

		const selectedText = getSelectedText() || undefined;
		const surroundingText = getSurroundingText() || undefined;

		inputText = '';
		clearSelection();

		await sendChatMessage(paperId, chatId, content, selectedText, surroundingText);
	}
</script>

<div class="input-area">
	{#if getSelectedText()}
		<div class="selection-badge">
			<span class="badge-label">Selected text:</span>
			<span class="badge-text">{getSelectedText().slice(0, 80)}{getSelectedText().length > 80 ? '...' : ''}</span>
			<button class="badge-clear" onclick={clearSelection} aria-label="Clear selection">&times;</button>
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
			Send
		</button>
	</div>
</div>

<style>
	.input-area {
		border-top: 1px solid #ddd;
		padding: 0.75rem;
	}

	.selection-badge {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.4rem 0.6rem;
		margin-bottom: 0.5rem;
		background: #fff3cd;
		border-radius: 4px;
		font-size: 0.8rem;
	}

	.badge-label {
		font-weight: 600;
		white-space: nowrap;
	}

	.badge-text {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: #555;
	}

	.badge-clear {
		border: none;
		background: none;
		cursor: pointer;
		font-size: 1rem;
		color: #999;
		padding: 0 0.25rem;
	}

	.badge-clear:hover {
		color: #333;
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
</style>
