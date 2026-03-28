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
vi.mock('$lib/uuid', () => ({
	generateId: () => 'test-uuid'
}));

import { sendMessage } from '$lib/api';
import { sendChatMessage } from '$lib/chat.svelte';

const paperId = 'paper-1';
const chatId = 'chat-1';

beforeEach(() => {
	vi.restoreAllMocks();
});

describe('sendChatMessage', () => {
	it('passes context to sendMessage when provided', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await sendChatMessage(paperId, chatId, 'Hello', {
			selectedText: 'quoted text',
			surroundingText: 'page content'
		});

		expect(sendMessage).toHaveBeenCalledWith(
			paperId,
			chatId,
			'Hello',
			expect.any(Function),
			expect.any(Function),
			expect.any(Function),
			{ selectedText: 'quoted text', surroundingText: 'page content' }
		);
	});

	it('passes undefined context when not provided', async () => {
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
			undefined
		);
	});

	it('includes selected_text on user message when provided', async () => {
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await sendChatMessage(paperId, chatId, 'Hello', {
			selectedText: 'quoted text',
			surroundingText: 'page content'
		});

		// Verify the user message was created with selected_text
		const { getMessages } = await import('$lib/chat.svelte');
		const messages = getMessages();
		const userMsg = messages.find(m => m.role === 'user');
		expect(userMsg).toBeDefined();
		expect(userMsg!.selected_text).toBe('quoted text');
	});
});
