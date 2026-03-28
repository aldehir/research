<script lang="ts">
	import { getMessages, getStreamingContent, getIsStreaming, getActiveToolCall } from '$lib/chat.svelte';
	import { tick } from 'svelte';

	let container: HTMLDivElement | undefined = $state();

	async function scrollToBottom() {
		await tick();
		if (container) {
			container.scrollTop = container.scrollHeight;
		}
	}

	const toolLabels: Record<string, string> = {
		search_pdf: 'Searching PDF...',
		read_page: 'Reading page...',
		go_to_page: 'Navigating...',
	};

	function toolLabel(name: string): string {
		return toolLabels[name] ?? 'Using tool...';
	}

	$effect(() => {
		getMessages();
		getStreamingContent();
		getActiveToolCall();
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
				{#if getActiveToolCall()}
					<div class="content tool-activity">{toolLabel(getActiveToolCall()!.name)}</div>
				{:else}
					<div class="content thinking">Thinking...</div>
				{/if}
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

	.tool-activity {
		color: #5f6368;
		font-style: italic;
		font-size: 0.85rem;
	}

</style>
