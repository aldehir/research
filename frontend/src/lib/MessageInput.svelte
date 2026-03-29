<script lang="ts">
	import { getIsStreaming, sendChatMessage, getActiveSessionId } from '$lib/chat.svelte';
	import { getCurrentPage } from '$lib/pdf-context.svelte';
	import { Icon, Send } from '$lib/icons';

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

		const currentPage = getCurrentPage();

		inputText = '';

		await sendChatMessage(paperId, chatId, content, currentPage);
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

	@media (max-width: 1023px) {
		.send-btn {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
