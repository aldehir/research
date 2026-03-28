<script lang="ts">
	import { getMessages, getStreamingContent, getIsStreaming } from '$lib/chat.svelte';
	import { tick } from 'svelte';

	let container: HTMLDivElement | undefined = $state();

	async function scrollToBottom() {
		await tick();
		if (container) {
			container.scrollTop = container.scrollHeight;
		}
	}

	$effect(() => {
		getMessages();
		getStreamingContent();
		scrollToBottom();
	});
</script>

<div class="thread" bind:this={container}>
	{#if getMessages().length === 0 && !getIsStreaming()}
		<p class="empty">Send a message to start the conversation</p>
	{:else}
		{#each getMessages() as message (message.id)}
			<div class="message {message.role}">
				<div class="role-label">{message.role === 'user' ? 'You' : 'Assistant'}</div>
				{#if message.selected_text}
					<blockquote class="selected-context">{message.selected_text}</blockquote>
				{/if}
				<div class="content">{message.content}</div>
			</div>
		{/each}
		{#if getIsStreaming() && getStreamingContent()}
			<div class="message assistant">
				<div class="role-label">Assistant</div>
				<div class="content">{getStreamingContent()}</div>
			</div>
		{/if}
		{#if getIsStreaming() && !getStreamingContent()}
			<div class="message assistant">
				<div class="role-label">Assistant</div>
				<div class="content thinking">Thinking...</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.thread {
		flex: 1;
		overflow-y: auto;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.empty {
		color: #888;
		text-align: center;
		margin-top: 2rem;
		font-size: 0.9rem;
	}

	.message {
		padding: 0.75rem;
		border-radius: 8px;
		max-width: 90%;
	}

	.message.user {
		background: #e8f0fe;
		align-self: flex-end;
	}

	.message.assistant {
		background: #f5f5f5;
		align-self: flex-start;
	}

	.role-label {
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		color: #666;
		margin-bottom: 0.25rem;
	}

	.content {
		font-size: 0.9rem;
		line-height: 1.5;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.thinking {
		color: #888;
		font-style: italic;
	}

	.selected-context {
		margin: 0 0 0.5rem 0;
		padding: 0.5rem;
		border-left: 3px solid #aaa;
		background: rgba(0, 0, 0, 0.05);
		font-size: 0.8rem;
		color: #555;
	}
</style>
