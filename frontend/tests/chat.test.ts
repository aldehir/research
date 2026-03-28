import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
	listChatSessions,
	createChatSession,
	getChatSession,
	deleteChatSession,
	sendMessage
} from '$lib/api';

const paperId = '123e4567-e89b-12d3-a456-426614174000';
const chatId = '223e4567-e89b-12d3-a456-426614174001';

const mockSession = {
	id: chatId,
	paper_id: paperId,
	title: 'Test Chat',
	created_at: '2026-01-01T00:00:00Z'
};

const mockMessage = {
	id: '333e4567-e89b-12d3-a456-426614174002',
	chat_session_id: chatId,
	role: 'user' as const,
	content: 'Hello',
	created_at: '2026-01-01T00:00:00Z'
};

beforeEach(() => {
	vi.restoreAllMocks();
});

describe('listChatSessions', () => {
	it('fetches sessions for a paper', async () => {
		const sessions = [mockSession];
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(sessions)
		}));

		const result = await listChatSessions(paperId);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${paperId}/chats`);
		expect(result).toEqual(sessions);
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'not found' })
		}));

		await expect(listChatSessions(paperId)).rejects.toThrow('not found');
	});
});

describe('createChatSession', () => {
	it('creates a session with title', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockSession)
		}));

		const result = await createChatSession(paperId, 'Test Chat');

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${paperId}/chats`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ title: 'Test Chat' })
		});
		expect(result).toEqual(mockSession);
	});

	it('creates a session without title', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockSession)
		}));

		await createChatSession(paperId);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${paperId}/chats`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({})
		});
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'bad request' })
		}));

		await expect(createChatSession(paperId)).rejects.toThrow('bad request');
	});
});

describe('getChatSession', () => {
	it('fetches a session with messages', async () => {
		const sessionWithMessages = { ...mockSession, messages: [mockMessage] };
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(sessionWithMessages)
		}));

		const result = await getChatSession(paperId, chatId);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${paperId}/chats/${chatId}`);
		expect(result).toEqual(sessionWithMessages);
		expect(result.messages).toHaveLength(1);
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'not found' })
		}));

		await expect(getChatSession(paperId, chatId)).rejects.toThrow('not found');
	});
});

describe('deleteChatSession', () => {
	it('sends DELETE request', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true
		}));

		await deleteChatSession(paperId, chatId);

		expect(fetch).toHaveBeenCalledWith(`/api/papers/${paperId}/chats/${chatId}`, {
			method: 'DELETE'
		});
	});

	it('throws on error response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'forbidden' })
		}));

		await expect(deleteChatSession(paperId, chatId)).rejects.toThrow('forbidden');
	});
});

describe('sendMessage', () => {
	function makeSSEStream(events: string[]): ReadableStream<Uint8Array> {
		const encoder = new TextEncoder();
		const data = events.join('\n') + '\n';
		return new ReadableStream({
			start(controller) {
				controller.enqueue(encoder.encode(data));
				controller.close();
			}
		});
	}

	it('parses SSE delta and done events', async () => {
		const stream = makeSSEStream([
			'data: {"type":"delta","text":"Hello"}',
			'data: {"type":"delta","text":" world"}',
			'data: {"type":"done"}'
		]);

		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			body: stream
		}));

		const deltas: string[] = [];
		const onDelta = (text: string) => deltas.push(text);
		const onDone = vi.fn();
		const onError = vi.fn();

		await sendMessage(paperId, chatId, 'Hi', undefined, undefined, onDelta, onDone, onError);

		expect(deltas).toEqual(['Hello', ' world']);
		expect(onDone).toHaveBeenCalledOnce();
		expect(onError).not.toHaveBeenCalled();
	});

	it('sends selected_text and surrounding_text when provided', async () => {
		const stream = makeSSEStream([
			'data: {"type":"done"}'
		]);

		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			body: stream
		}));

		await sendMessage(
			paperId, chatId, 'explain this', 'some text', 'context around',
			vi.fn(), vi.fn(), vi.fn()
		);

		expect(fetch).toHaveBeenCalledWith(
			`/api/papers/${paperId}/chats/${chatId}/messages`,
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					content: 'explain this',
					selected_text: 'some text',
					surrounding_text: 'context around'
				})
			}
		);
	});

	it('calls onError when response is not ok', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({ error: 'bad request' })
		}));

		const onDelta = vi.fn();
		const onDone = vi.fn();
		const onError = vi.fn();

		await sendMessage(paperId, chatId, 'Hi', undefined, undefined, onDelta, onDone, onError);

		expect(onError).toHaveBeenCalledWith('bad request');
		expect(onDone).not.toHaveBeenCalled();
	});

	it('calls onError on network failure', async () => {
		vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

		const onError = vi.fn();

		await sendMessage(paperId, chatId, 'Hi', undefined, undefined, vi.fn(), vi.fn(), onError);

		expect(onError).toHaveBeenCalledWith('Network error');
	});

	it('handles chunked SSE data across multiple reads', async () => {
		const encoder = new TextEncoder();
		const stream = new ReadableStream<Uint8Array>({
			start(controller) {
				// Split across chunk boundary
				controller.enqueue(encoder.encode('data: {"type":"del'));
				controller.enqueue(encoder.encode('ta","text":"hi"}\n'));
				controller.enqueue(encoder.encode('data: {"type":"done"}\n'));
				controller.close();
			}
		});

		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			body: stream
		}));

		const deltas: string[] = [];
		const onDone = vi.fn();

		await sendMessage(paperId, chatId, 'Hi', undefined, undefined,
			(t) => deltas.push(t), onDone, vi.fn());

		expect(deltas).toEqual(['hi']);
		expect(onDone).toHaveBeenCalledOnce();
	});
});
