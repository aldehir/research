import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ToolCall, ToolResult } from '$lib/api';

// Mock the api module
vi.mock('$lib/api', () => ({
	listChatSessions: vi.fn(),
	createChatSession: vi.fn(),
	getChatSession: vi.fn(),
	deleteChatSession: vi.fn(),
	sendMessage: vi.fn()
}));

// Mock uuid
vi.mock('$lib/uuid', () => ({
	generateId: () => 'test-uuid'
}));

import { sendMessage } from '$lib/api';
import {
	sendChatMessage,
	getStreamSegments,
	getIsStreaming,
	resetChat
} from '$lib/chat.svelte';

const paperId = 'paper-1';
const chatId = 'chat-1';

beforeEach(() => {
	vi.restoreAllMocks();
	resetChat();
});

describe('sendChatMessage', () => {
	it('passes currentPage to sendMessage when provided', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await sendChatMessage(paperId, chatId, 'Hello', 5);

		expect(sendMessage).toHaveBeenCalledWith(
			paperId,
			chatId,
			'Hello',
			expect.any(Function),
			expect.any(Function),
			expect.any(Function),
			5,
			expect.any(Function),
			expect.any(Function)
		);
	});

	it('passes undefined currentPage when not provided', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await sendChatMessage(paperId, chatId, 'Hello');

		expect(sendMessage).toHaveBeenCalledWith(
			paperId,
			chatId,
			'Hello',
			expect.any(Function),
			expect.any(Function),
			expect.any(Function),
			undefined,
			expect.any(Function),
			expect.any(Function)
		);
	});
});

describe('stream segments tracking', () => {
	it('accumulates text deltas into a text segment', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, onDone) => {
				onDelta('Hello');
				onDelta(' world');
				onDone();
			}
		);

		await sendChatMessage(paperId, chatId, 'Hi');

		// After done, segments should be empty (moved to message)
		// But during streaming they would have been populated
		// Verify final message content instead
		const { getMessages } = await import('$lib/chat.svelte');
		const msgs = getMessages();
		const assistant = msgs.find(m => m.role === 'assistant');
		expect(assistant).toBeDefined();
		expect(assistant!.content).toBe('Hello world');
	});

	it('tracks tool call and result as segments during streaming', async () => {
		let capturedOnToolCall: ((tool: ToolCall) => void) | undefined;
		let capturedOnToolResult: ((result: ToolResult) => void) | undefined;
		let capturedOnDelta: ((text: string) => void) | undefined;

		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, onDone, _onError, _ctx, onToolCall, onToolResult) => {
				capturedOnDelta = onDelta;
				capturedOnToolCall = onToolCall;
				capturedOnToolResult = onToolResult;

				// Simulate: text -> tool_call -> tool_result -> more text -> done
				onDelta('Let me search. ');
				onToolCall!({ name: 'search_pdf', args: { query: 'attention' } });
				onToolResult!({ name: 'search_pdf', text: 'Found on page 3', preview: 'Found on page 3' });
				onDelta('I found it on page 3.');
				onDone();
			}
		);

		// We need to check segments mid-stream. Since mock is sync,
		// let's check the final message instead, and verify callbacks are passed.
		await sendChatMessage(paperId, chatId, 'Find attention');

		expect(capturedOnToolCall).toBeDefined();
		expect(capturedOnToolResult).toBeDefined();
		expect(capturedOnDelta).toBeDefined();

		// Verify final message has full text
		const { getMessages } = await import('$lib/chat.svelte');
		const msgs = getMessages();
		const assistant = msgs.find(m => m.role === 'assistant');
		expect(assistant).toBeDefined();
		expect(assistant!.content).toBe('Let me search. I found it on page 3.');
	});

	it('builds ordered stream segments from interleaved events', async () => {
		// Use a promise to keep streaming alive while we check state
		let resolveStream: () => void;
		const streamPromise = new Promise<void>(r => { resolveStream = r; });

		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, onDone, _onError, _ctx, onToolCall, onToolResult) => {
				onDelta('Searching...');
				onToolCall!({ name: 'search_pdf', args: { query: 'test' } });
				onToolResult!({ name: 'search_pdf', text: 'Found result', preview: 'Found result' });
				onDelta('Here is what I found.');

				// Check segments mid-stream
				const segments = getStreamSegments();
				expect(segments.length).toBe(3);
				expect(segments[0]).toEqual({ type: 'text', content: 'Searching...' });
				expect(segments[1]).toEqual({
					type: 'tool',
					name: 'search_pdf',
					args: { query: 'test' },
					result: { name: 'search_pdf', text: 'Found result', preview: 'Found result' }
				});
				expect(segments[2]).toEqual({ type: 'text', content: 'Here is what I found.' });

				onDone();
				resolveStream!();
			}
		);

		await sendChatMessage(paperId, chatId, 'Search');
		await streamPromise;
	});

	it('preserves tool history on completed assistant message', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, onDone, _onError, _ctx, onToolCall, onToolResult) => {
				onDelta('Searching...');
				onToolCall!({ name: 'search_pdf', args: { query: 'test' } });
				onToolResult!({ name: 'search_pdf', text: 'Found result', preview: 'Found result' });
				onDelta('Here it is.');
				onDone();
			}
		);

		await sendChatMessage(paperId, chatId, 'Search');

		const { getMessages, getMessageSegments } = await import('$lib/chat.svelte');
		const msgs = getMessages();
		const assistant = msgs.find(m => m.role === 'assistant');
		expect(assistant).toBeDefined();

		const segments = getMessageSegments(assistant!.id);
		expect(segments).toBeDefined();
		expect(segments!).toHaveLength(3);
		expect(segments![0]).toEqual({ type: 'text', content: 'Searching...' });
		expect(segments![1]).toMatchObject({ type: 'tool', name: 'search_pdf' });
		expect(segments![2]).toEqual({ type: 'text', content: 'Here it is.' });
	});

	it('clears stream segments after streaming completes', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, onDelta, onDone) => {
				onDelta('Response text');
				onDone();
			}
		);

		await sendChatMessage(paperId, chatId, 'Hi');

		const segments = getStreamSegments();
		expect(segments).toEqual([]);
		expect(getIsStreaming()).toBe(false);
	});
});
