<script lang="ts">
	import { getSessions, getActiveSessionId, selectSession, deleteSession, createSession } from '$lib/chat.svelte';

	interface Props {
		paperId: string;
	}

	let { paperId }: Props = $props();

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric'
		});
	}

	function handleSelect(chatId: string) {
		selectSession(paperId, chatId);
	}

	async function handleDelete(event: Event, chatId: string) {
		event.stopPropagation();
		await deleteSession(paperId, chatId);
	}

	function handleNew() {
		createSession(paperId);
	}
</script>

<div class="session-list">
	<button class="new-chat-btn" onclick={handleNew}>+ New Chat</button>
	{#if getSessions().length === 0}
		<p class="empty">No conversations yet</p>
	{:else}
		<ul>
			{#each getSessions() as session (session.id)}
				<li>
					<button
						class="session-item"
						class:active={getActiveSessionId() === session.id}
						onclick={() => handleSelect(session.id)}
					>
						<span class="session-title">{session.title}</span>
						<span class="session-date">{formatDate(session.created_at)}</span>
					</button>
					<button
						class="delete-btn"
						onclick={(e) => handleDelete(e, session.id)}
						aria-label="Delete chat"
					>
						&times;
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.session-list {
		border-bottom: 1px solid #ddd;
		max-height: 200px;
		overflow-y: auto;
	}

	.new-chat-btn {
		width: 100%;
		padding: 0.5rem 1rem;
		border: none;
		background: #e8f0fe;
		cursor: pointer;
		font-weight: 500;
		text-align: left;
	}

	.new-chat-btn:hover {
		background: #d2e3fc;
	}

	.empty {
		color: #888;
		text-align: center;
		padding: 1rem;
		font-size: 0.85rem;
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	li {
		display: flex;
		align-items: center;
		border-bottom: 1px solid #eee;
	}

	.session-item {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		padding: 0.5rem 1rem;
		border: none;
		background: none;
		cursor: pointer;
		text-align: left;
	}

	.session-item:hover {
		background: #f5f5f5;
	}

	.session-item.active {
		background: #e8f0fe;
	}

	.session-title {
		font-size: 0.85rem;
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.session-date {
		font-size: 0.75rem;
		color: #666;
	}

	.delete-btn {
		padding: 0.25rem 0.5rem;
		border: none;
		background: none;
		cursor: pointer;
		color: #999;
		font-size: 1.1rem;
		line-height: 1;
	}

	.delete-btn:hover {
		color: #e00;
	}

	@media (max-width: 1023px) {
		.new-chat-btn {
			min-height: 44px;
			display: flex;
			align-items: center;
		}

		.session-item {
			min-height: 44px;
			padding: 0.5rem 1rem;
		}

		.delete-btn {
			min-width: 44px;
			min-height: 44px;
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}
</style>
