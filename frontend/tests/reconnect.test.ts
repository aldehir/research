import { describe, it, expect, vi, beforeEach } from 'vitest';
import { reconnectStream } from '$lib/api';
import type { SSECallbacks } from '$lib/api';

beforeEach(() => {
	vi.restoreAllMocks();
});

function makeCallbacks(): SSECallbacks & { calls: string[] } {
	const calls: string[] = [];
	return {
		calls,
		onDelta: (text: string) => calls.push(`delta:${text}`),
		onDone: () => calls.push('done'),
		onError: (error: string) => calls.push(`error:${error}`)
	};
}

function mockSSEResponse(events: string[]): Response {
	const sseText = events.map(e => `data: ${e}\n\n`).join('');
	const encoder = new TextEncoder();
	const stream = new ReadableStream({
		start(controller) {
			controller.enqueue(encoder.encode(sseText));
			controller.close();
		}
	});
	return new Response(stream, {
		status: 200,
		headers: { 'Content-Type': 'text/event-stream' }
	});
}

describe('reconnectStream', () => {
	it('returns false on 404', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 404,
			json: () => Promise.resolve({ error: 'no active stream' })
		}));

		const cb = makeCallbacks();
		const result = await reconnectStream('paper-1', 'chat-1', cb);

		expect(result).toBe(false);
		expect(cb.calls).toEqual([]);
	});

	it('returns false on fetch error', async () => {
		vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network error')));

		const cb = makeCallbacks();
		const result = await reconnectStream('paper-1', 'chat-1', cb);

		expect(result).toBe(false);
	});

	it('returns true and replays events from active stream', async () => {
		const response = mockSSEResponse([
			'{"type":"delta","text":"Hello"}',
			'{"type":"delta","text":" world"}',
			'{"type":"done"}'
		]);
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response));

		const cb = makeCallbacks();
		const result = await reconnectStream('paper-1', 'chat-1', cb);

		expect(result).toBe(true);
		expect(fetch).toHaveBeenCalledWith(
			'/api/papers/paper-1/chats/chat-1/stream',
			expect.objectContaining({})
		);

		// Wait for async stream consumption
		await new Promise(resolve => setTimeout(resolve, 50));

		expect(cb.calls).toEqual(['delta:Hello', 'delta: world', 'done']);
	});

	it('fetches the correct endpoint', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: false,
			status: 404
		}));

		await reconnectStream('p-42', 'c-99', makeCallbacks());

		expect(fetch).toHaveBeenCalledWith(
			'/api/papers/p-42/chats/c-99/stream',
			expect.objectContaining({})
		);
	});
});
