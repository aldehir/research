<script lang="ts">
	import MessageThread from '$lib/MessageThread.svelte';
	import MessageInput from '$lib/MessageInput.svelte';
	import { loadSessions, getSessions, getActiveSessionId, selectSession, deleteSession, createSession, resetChat } from '$lib/chat.svelte';
	import { getIsMobile } from '$lib/mobile-layout.svelte';
	import { Icon, PanelRightOpen, PanelRightClose, Plus, ChevronDown, X } from '$lib/icons';

	interface Props {
		paperId: string;
		chatWidth?: number;
	}

	let { paperId, chatWidth }: Props = $props();
	let collapsed = $state(false);
	let dropdownOpen = $state(false);
	let previousPaperId = $state('');

	let activeSession = $derived(getSessions().find(s => s.id === getActiveSessionId()));

	$effect(() => {
		if (paperId !== previousPaperId) {
			previousPaperId = paperId;
			resetChat();
			loadSessions(paperId);
		}
	});

	function handleNew() {
		createSession(paperId);
		dropdownOpen = false;
	}

	function handleSelect(chatId: string) {
		selectSession(paperId, chatId);
		dropdownOpen = false;
	}

	async function handleDelete(event: Event, chatId: string) {
		event.stopPropagation();
		await deleteSession(paperId, chatId);
	}

	function handleBackdropClick() {
		dropdownOpen = false;
	}
</script>

{#if !getIsMobile() && collapsed}
	<div class="chat-collapsed">
		<button class="toggle-btn" onclick={() => collapsed = false} aria-label="Open chat">
			<Icon d={PanelRightOpen} size={18} />
		</button>
	</div>
{:else}
	<div class="chat-panel" style:width={chatWidth ? `${chatWidth}px` : undefined}>
		<div class="chat-header">
			<div class="session-picker">
				<button
					class="picker-btn"
					onclick={() => dropdownOpen = !dropdownOpen}
					aria-label="Switch chat session"
				>
					<span class="picker-label">{activeSession?.title ?? 'No chat selected'}</span>
					<Icon d={ChevronDown} size={16} />
				</button>
				<button class="new-btn" onclick={handleNew} aria-label="New chat" title="New chat">
					<Icon d={Plus} size={18} />
				</button>
				{#if dropdownOpen}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div class="dropdown-backdrop" onclick={handleBackdropClick}></div>
					<div class="dropdown">
						{#if getSessions().length === 0}
							<p class="dropdown-empty">No conversations yet</p>
						{:else}
							{#each getSessions() as session (session.id)}
								<!-- svelte-ignore a11y_click_events_have_key_events -->
								<!-- svelte-ignore a11y_no_static_element_interactions -->
								<div
									class="dropdown-item"
									class:active={getActiveSessionId() === session.id}
									onclick={() => handleSelect(session.id)}
								>
									<span class="dropdown-item-title">{session.title}</span>
									<button
										class="dropdown-item-delete"
										onclick={(e) => handleDelete(e, session.id)}
										aria-label="Delete chat"
									>
										<Icon d={X} size={14} />
									</button>
								</div>
							{/each}
						{/if}
					</div>
				{/if}
			</div>
			{#if !getIsMobile()}
				<button class="toggle-btn" onclick={() => collapsed = true} aria-label="Close chat">
					<Icon d={PanelRightClose} size={18} />
				</button>
			{/if}
		</div>

		{#if getActiveSessionId()}
			<MessageThread />
			<MessageInput {paperId} />
		{:else}
			<div class="no-session">
				<p>Start a conversation with <button class="inline-new" onclick={handleNew}><Icon d={Plus} size={14} /> New Chat</button></p>
			</div>
		{/if}
	</div>
{/if}

<style>
	.chat-panel {
		width: 360px;
		min-width: 0;
		border-left: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		background: var(--color-bg);
		height: 100%;
		position: relative;
		flex-shrink: 0;
	}

	.chat-collapsed {
		display: flex;
		align-items: flex-start;
		border-left: 1px solid var(--color-border);
		padding-top: 0.5rem;
	}

	.chat-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 0.5rem 0.5rem 0.75rem;
		border-bottom: 1px solid var(--color-border);
		gap: 0.25rem;
	}

	.session-picker {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		flex: 1;
		min-width: 0;
		position: relative;
	}

	.picker-btn {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		flex: 1;
		min-width: 0;
		height: var(--btn-height-md);
		padding: 0 0.5rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		background: var(--color-bg);
		color: var(--color-text);
		cursor: pointer;
		font-size: 0.85rem;
		font-weight: 500;
	}

	.picker-btn:hover {
		background: var(--color-surface-hover);
	}

	.picker-label {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-align: left;
	}

	.new-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: var(--btn-height-md);
		height: var(--btn-height-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		background: var(--color-bg);
		color: var(--color-text-secondary);
		cursor: pointer;
		flex-shrink: 0;
	}

	.new-btn:hover {
		background: var(--color-surface-hover);
		color: var(--color-text);
	}

	.toggle-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		background: none;
		cursor: pointer;
		color: var(--color-text-secondary);
		padding: 0.25rem;
		border-radius: var(--radius-sm);
		flex-shrink: 0;
	}

	.toggle-btn:hover {
		color: var(--color-text);
		background: var(--color-surface-hover);
	}

	/* Dropdown */
	.dropdown-backdrop {
		position: fixed;
		inset: 0;
		z-index: 199;
	}

	.dropdown {
		position: absolute;
		top: 100%;
		left: 0;
		right: 0;
		z-index: 200;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		box-shadow: 0 4px 16px var(--color-shadow);
		max-height: 240px;
		overflow-y: auto;
		margin-top: -1px;
	}

	.dropdown-empty {
		color: var(--color-text-tertiary);
		text-align: center;
		padding: 1rem;
		margin: 0;
		font-size: 0.85rem;
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		width: 100%;
		padding: 0.5rem 0.75rem;
		border: none;
		background: none;
		cursor: pointer;
		text-align: left;
		color: var(--color-text);
		font-size: 0.85rem;
		gap: 0.5rem;
	}

	.dropdown-item:hover {
		background: var(--color-surface-hover);
	}

	.dropdown-item.active {
		background: var(--color-surface-active);
	}

	.dropdown-item + .dropdown-item {
		border-top: 1px solid var(--color-border);
	}

	.dropdown-item-title {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.dropdown-item-delete {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border: none;
		background: none;
		cursor: pointer;
		color: var(--color-text-tertiary);
		border-radius: var(--radius-sm);
		flex-shrink: 0;
	}

	.dropdown-item-delete:hover {
		color: var(--color-danger);
		background: var(--color-surface-hover);
	}

	/* Empty state */
	.no-session {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--color-text-tertiary);
		font-size: 0.9rem;
	}

	.no-session p {
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}

	.inline-new {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		border: none;
		background: none;
		color: var(--color-primary);
		cursor: pointer;
		font-size: inherit;
		font-weight: 500;
		padding: 0;
	}

	.inline-new:hover {
		text-decoration: underline;
	}

	@media (max-width: 1023px) {
		.toggle-btn {
			min-width: 44px;
			min-height: 44px;
		}

		.picker-btn {
			min-height: 44px;
		}

		.new-btn {
			min-width: 44px;
			min-height: 44px;
		}

		.dropdown-item {
			min-height: 44px;
		}

		.dropdown-item-delete {
			min-width: 44px;
			min-height: 44px;
		}
	}
</style>
