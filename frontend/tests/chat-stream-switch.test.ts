import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the api module
vi.mock('$lib/api', () => ({
	listChatSessions: vi.fn(),
	createChatSession: vi.fn(),
	getChatSession: vi.fn(),
	deleteChatSession: vi.fn(),
	sendMessage: vi.fn()
}));

// Mock uuid
let uuidCounter = 0;
vi.mock('$lib/uuid', () => ({
	generateId: () => `test-uuid-${++uuidCounter}`
}));

// Mock pdf-navigate
vi.mock('$lib/pdf-navigate.svelte', () => ({
	requestGoToPage: vi.fn()
}));

import { getChatSession, sendMessage } from '$lib/api';
import {
	sendChatMessage,
	selectSession,
	createSession,
	getMessages,
	getIsStreaming,
	getStreamingContent,
	getStreamSegments,
	resetChat
} from '$lib/chat.svelte';

const paperId = 'paper-1';
const chatA = 'chat-a';
const chatB = 'chat-b';

beforeEach(() => {
	vi.restoreAllMocks();
	uuidCounter = 0;
	resetChat();
});

describe('switching chat during streaming', () => {
	it('aborts active stream and clears streaming state on session switch', async () => {
		// Start streaming in chat A — keep the stream alive via a deferred promise
		let resolveStream!: () => void;
		let capturedSignal: AbortSignal | undefined;

		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, _onDone, _onError, _ctx, _onTool, _onResult, _att, signal) => {
				capturedSignal = signal;
				onDelta('Hello from A');
				// Stream stays alive until resolved
				await new Promise<void>(r => { resolveStream = r; });
			}
		);

		const streamPromise = sendChatMessage(paperId, chatA, 'Hi');

		// Verify streaming state is active
		expect(getIsStreaming()).toBe(true);
		expect(getStreamingContent()).toBe('Hello from A');

		// Switch to chat B
		vi.mocked(getChatSession).mockResolvedValue({
			id: chatB,
			paper_id: paperId,
			title: 'Chat B',
			created_at: '2026-01-01T00:00:00Z',
			messages: [{ id: 'msg-b1', chat_session_id: chatB, role: 'user', content: 'old msg', created_at: '2026-01-01T00:00:00Z' }]
		});

		await selectSession(paperId, chatB);

		// Streaming state should be cleared — no stale content from chat A
		expect(getIsStreaming()).toBe(false);
		expect(getStreamingContent()).toBe('');
		expect(getStreamSegments()).toEqual([]);

		// The abort signal should have been triggered
		expect(capturedSignal?.aborted).toBe(true);

		// Chat B's messages should be loaded, not contaminated
		const msgs = getMessages();
		expect(msgs).toHaveLength(1);
		expect(msgs[0].content).toBe('old msg');

		resolveStream();
		await streamPromise;
	});

	it('onDone no-ops if chat was switched mid-stream', async () => {
		let capturedOnDone!: () => void;

		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, onDone, _onError, _ctx, _onTool, _onResult, _att, _signal) => {
				capturedOnDone = onDone;
				onDelta('partial response');
				// Stream stays alive — we'll call onDone manually after switching
				await new Promise<void>(() => {});
			}
		);

		// Fire and forget — stream will never naturally resolve
		sendChatMessage(paperId, chatA, 'Hi');

		// Switch to chat B
		vi.mocked(getChatSession).mockResolvedValue({
			id: chatB,
			paper_id: paperId,
			title: 'Chat B',
			created_at: '2026-01-01T00:00:00Z',
			messages: []
		});
		await selectSession(paperId, chatB);

		// Simulate the original stream completing after switch
		capturedOnDone();

		// The assistant message from chat A should NOT appear in chat B's messages
		const msgs = getMessages();
		const assistantMsg = msgs.find(m => m.role === 'assistant');
		expect(assistantMsg).toBeUndefined();

		// Streaming state should remain cleared
		expect(getIsStreaming()).toBe(false);
	});

	it('creating new chat during streaming aborts stream and clears state', async () => {
		let capturedSignal: AbortSignal | undefined;

		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, _onDone, _onError, _ctx, _onTool, _onResult, _att, signal) => {
				capturedSignal = signal;
				onDelta('Hello from A');
				await new Promise<void>(() => {});
			}
		);

		// Send a message in an existing chat, then create a new chat while streaming
		sendChatMessage(paperId, chatA, 'Hello');

		// Verify streaming is active
		expect(getIsStreaming()).toBe(true);
		expect(getStreamingContent()).toBe('Hello from A');

		// Create a new chat — should abort the stream
		await createSession(paperId);

		expect(getIsStreaming()).toBe(false);
		expect(getStreamingContent()).toBe('');
		expect(getStreamSegments()).toEqual([]);
		expect(capturedSignal?.aborted).toBe(true);

		// New chat should have no messages
		expect(getMessages()).toEqual([]);
	});

	it('deltas after session switch do not pollute new session', async () => {
		let capturedOnDelta!: (text: string) => void;

		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, _onDone, _onError, _ctx, _onTool, _onResult, _att, _signal) => {
				capturedOnDelta = onDelta;
				onDelta('start');
				await new Promise<void>(() => {});
			}
		);

		sendChatMessage(paperId, chatA, 'Hi');

		// Switch to chat B
		vi.mocked(getChatSession).mockResolvedValue({
			id: chatB,
			paper_id: paperId,
			title: 'Chat B',
			created_at: '2026-01-01T00:00:00Z',
			messages: []
		});
		await selectSession(paperId, chatB);

		// Late delta arrives from chat A's stream
		capturedOnDelta(' late text');

		// Chat B's streaming state should not be contaminated
		expect(getStreamingContent()).toBe('');
		expect(getStreamSegments()).toEqual([]);
		expect(getIsStreaming()).toBe(false);
	});
});
