<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';

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

		inputText = '';

		await sendChatMessage(paperId, chatId, content);
	}
</script>

<div class="input-area">
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
