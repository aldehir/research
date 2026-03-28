<script lang="ts">
	import { getMessages, getStreamSegments, getStreamingContent, getIsStreaming, getMessageSegments } from '$lib/chat.svelte';
	import type { StreamSegment } from '$lib/chat.svelte';
	import { formatToolLabel, formatToolArgs } from '$lib/tool-display';
	import MarkdownRenderer from '$lib/MarkdownRenderer.svelte';
	import { tick } from 'svelte';

	let container: HTMLDivElement | undefined = $state();
	let expandedTools = $state(new Set<string>());
	const messages = $derived(getMessages());
	const streaming = $derived(getIsStreaming());
	const segments = $derived(getStreamSegments());

	async function scrollToBottom() {
		await tick();
		if (container) {
			container.scrollTop = container.scrollHeight;
		}
	}

	function toggleTool(key: string) {
		const next = new Set(expandedTools);
		if (next.has(key)) {
			next.delete(key);
		} else {
			next.add(key);
		}
		expandedTools = next;
	}

	function segmentKey(messageId: string, index: number): string {
		return `${messageId}-${index}`;
	}

	type GroupedSegment =
		| { type: 'markdown'; content: string }
		| { type: 'tool'; segment: StreamSegment & { type: 'tool' }; index: number };

	function groupSegments(segs: StreamSegment[]): GroupedSegment[] {
		const groups: GroupedSegment[] = [];
		let textBuffer = '';

		for (let i = 0; i < segs.length; i++) {
			const seg = segs[i];
			if (seg.type === 'text') {
				textBuffer += seg.content;
			} else {
				if (textBuffer) {
					groups.push({ type: 'markdown', content: textBuffer });
					textBuffer = '';
				}
				groups.push({ type: 'tool', segment: seg, index: i });
			}
		}
		if (textBuffer) {
			groups.push({ type: 'markdown', content: textBuffer });
		}
		return groups;
	}

	$effect(() => {
		messages;
		getStreamingContent();
		scrollToBottom();
	});
</script>

{#snippet toolChip(segment: StreamSegment & { type: 'tool' }, key: string)}
	<div class="tool-chip">
		<button class="tool-chip-header" onclick={() => toggleTool(key)}>
			<span class="tool-chip-icon">{expandedTools.has(key) ? '▾' : '▸'}</span>
			<span class="tool-chip-label">{formatToolLabel(segment.name)}</span>
			<span class="tool-chip-args">{formatToolArgs(segment.name, segment.args)}</span>
			{#if !segment.result}
				<span class="tool-chip-spinner"></span>
			{/if}
		</button>
		{#if expandedTools.has(key) && segment.result}
			<div class="tool-result-popout">
				{#if segment.result.content_type === 'image' && segment.result.image_data}
					<img class="tool-result-image" src="data:image/png;base64,{segment.result.image_data}" alt="Page snapshot" />
				{:else}
					<pre class="tool-result-content">{segment.result.text}</pre>
				{/if}
			</div>
		{/if}
	</div>
{/snippet}

{#snippet segmentList(segs: StreamSegment[], messageId: string)}
	{#each groupSegments(segs) as group}
		{#if group.type === 'markdown'}
			<MarkdownRenderer content={group.content} />
		{:else}
			{@render toolChip(group.segment, segmentKey(messageId, group.index))}
		{/if}
	{/each}
{/snippet}

<div class="thread" bind:this={container}>
	{#if messages.length === 0 && !streaming}
		<p class="empty">Send a message to start the conversation</p>
	{:else}
		{#each messages as message (message.id)}
			<div class="message {message.role}">
				<div class="role-label">{message.role === 'user' ? 'You' : 'Assistant'}</div>
				{#if message.role === 'assistant' && getMessageSegments(message.id)}
					<div class="content assistant-content">
						{@render segmentList(getMessageSegments(message.id)!, message.id)}
					</div>
				{:else if message.role === 'assistant'}
					<div class="content assistant-content">
						<MarkdownRenderer content={message.content} />
					</div>
				{:else}
					<div class="content">{message.content}</div>
				{/if}
			</div>
		{/each}
		{#if streaming}
			{#if segments.length > 0}
				<div class="message assistant">
					<div class="role-label">Assistant</div>
					<div class="content assistant-content">
						{@render segmentList(segments, 'streaming')}
					</div>
				</div>
			{:else}
				<div class="message assistant">
					<div class="role-label">Assistant</div>
					<div class="content thinking">Thinking...</div>
				</div>
			{/if}
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

	.content.assistant-content {
		white-space: normal;
	}

	.thinking {
		color: #888;
		font-style: italic;
	}

	.tool-chip {
		display: block;
		margin: 0.4rem 0;
	}

	.tool-chip-header {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		padding: 0.25rem 0.5rem;
		background: #e3e8ef;
		border: 1px solid #cbd2dc;
		border-radius: 4px;
		font-size: 0.8rem;
		color: #444;
		cursor: pointer;
		font-family: inherit;
	}

	.tool-chip-header:hover {
		background: #d5dce6;
	}

	.tool-chip-icon {
		font-size: 0.7rem;
		width: 0.8rem;
	}

	.tool-chip-label {
		font-weight: 500;
	}

	.tool-chip-args {
		color: #666;
	}

	.tool-chip-spinner {
		display: inline-block;
		width: 0.7rem;
		height: 0.7rem;
		border: 1.5px solid #999;
		border-top-color: transparent;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.tool-result-popout {
		margin-top: 0.3rem;
		padding: 0.5rem;
		background: #fff;
		border: 1px solid #ddd;
		border-radius: 4px;
		max-height: 300px;
		overflow-y: auto;
	}

	.tool-result-content {
		margin: 0;
		font-size: 0.8rem;
		line-height: 1.4;
		white-space: pre-wrap;
		word-break: break-word;
		color: #333;
	}

	.tool-result-image {
		max-width: 100%;
		height: auto;
		border-radius: 4px;
	}
</style>
