<script lang="ts">
	import ChatSessionList from '$lib/ChatSessionList.svelte';
	import MessageThread from '$lib/MessageThread.svelte';
	import MessageInput from '$lib/MessageInput.svelte';
	import { loadSessions, getActiveSessionId, resetChat } from '$lib/chat.svelte';
	import { getIsMobile } from '$lib/mobile-layout.svelte';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();
	let collapsed = $state(false);
	let previousPaperId = $state('');

	$effect(() => {
		if (paperId !== previousPaperId) {
			previousPaperId = paperId;
			resetChat();
			loadSessions(paperId);
		}
	});
</script>

{#if !getIsMobile() && collapsed}
	<div class="chat-collapsed">
		<button class="toggle-btn" onclick={() => collapsed = false} aria-label="Open chat">
			&#x25C0;
		</button>
	</div>
{:else}
	<div class="chat-panel">
		<div class="chat-header">
			<h3>Chat</h3>
			{#if !getIsMobile()}
				<button class="toggle-btn" onclick={() => collapsed = true} aria-label="Close chat">
					&#x25B6;
				</button>
			{/if}
		</div>
		<ChatSessionList {paperId} />
		{#if getActiveSessionId()}
			<MessageThread />
			<MessageInput {paperId} />
		{:else}
			<div class="no-session">
				<p>Create or select a chat to begin</p>
			</div>
		{/if}
	</div>
{/if}

<style>
	.chat-panel {
		width: 360px;
		min-width: 300px;
		border-left: 1px solid #ddd;
		display: flex;
		flex-direction: column;
		background: #fff;
		height: 100%;
	}

	.chat-collapsed {
		display: flex;
		align-items: flex-start;
		border-left: 1px solid #ddd;
		padding-top: 0.5rem;
	}

	.chat-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #ddd;
	}

	.chat-header h3 {
		margin: 0;
		font-size: 1rem;
	}

	.toggle-btn {
		border: none;
		background: none;
		cursor: pointer;
		font-size: 0.9rem;
		color: #666;
		padding: 0.25rem 0.5rem;
	}

	.toggle-btn:hover {
		color: #333;
	}

	.no-session {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #888;
		font-size: 0.9rem;
	}

	@media (max-width: 1023px) {
		.toggle-btn {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
