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

// Mock pdf-navigate
vi.mock('$lib/pdf-navigate.svelte', () => ({
	requestGoToPage: vi.fn()
}));

import {
	createChatSession,
	deleteChatSession,
	getChatSession,
	listChatSessions,
	sendMessage
} from '$lib/api';
import {
	createSession,
	sendChatMessage,
	selectSession,
	deleteSession,
	loadSessions,
	getSessions,
	getActiveSessionId,
	getMessages,
	resetChat
} from '$lib/chat.svelte';

const paperId = 'paper-1';

beforeEach(() => {
	vi.restoreAllMocks();
	resetChat();
});

describe('draft session creation', () => {
	it('createSession does not call the API', async () => {
		await createSession(paperId);

		expect(createChatSession).not.toHaveBeenCalled();
	});

	it('createSession adds a session with draft- prefix ID', async () => {
		await createSession(paperId);

		const sessions = getSessions();
		expect(sessions).toHaveLength(1);
		expect(sessions[0].id).toMatch(/^draft-/);
		expect(getActiveSessionId()).toBe(sessions[0].id);
	});

	it('createSession reuses existing draft instead of creating another', async () => {
		await createSession(paperId);
		const firstDraftId = getActiveSessionId();

		await createSession(paperId);

		const sessions = getSessions();
		expect(sessions).toHaveLength(1);
		expect(getActiveSessionId()).toBe(firstDraftId);
	});
});

describe('draft session promotion on first message', () => {
	it('calls createChatSession API then sendMessage on first message', async () => {
		const realChatId = 'real-chat-id';
		vi.mocked(createChatSession).mockResolvedValue({
			id: realChatId,
			paper_id: paperId,
			title: 'New Chat',
			created_at: '2026-01-01T00:00:00Z'
		});
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await createSession(paperId);
		const draftId = getActiveSessionId()!;
		expect(draftId).toMatch(/^draft-/);

		await sendChatMessage(paperId, draftId, 'Hello');

		expect(createChatSession).toHaveBeenCalledWith(paperId);
		expect(sendMessage).toHaveBeenCalledWith(
			paperId,
			realChatId,
			'Hello',
			expect.any(Function),
			expect.any(Function),
			expect.any(Function),
			undefined,
			expect.any(Function),
			expect.any(Function),
			undefined,
			expect.any(AbortSignal)
		);
	});

	it('replaces draft ID with real ID in sessions list and activeSessionId', async () => {
		const realChatId = 'real-chat-id';
		vi.mocked(createChatSession).mockResolvedValue({
			id: realChatId,
			paper_id: paperId,
			title: 'New Chat',
			created_at: '2026-01-01T00:00:00Z'
		});
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await createSession(paperId);
		await sendChatMessage(paperId, getActiveSessionId()!, 'Hello');

		expect(getActiveSessionId()).toBe(realChatId);
		const sessions = getSessions();
		expect(sessions.find(s => s.id === realChatId)).toBeDefined();
		expect(sessions.find(s => s.id.startsWith('draft-'))).toBeUndefined();
	});

	it('updates chat_session_id on user message after promotion', async () => {
		const realChatId = 'real-chat-id';
		vi.mocked(createChatSession).mockResolvedValue({
			id: realChatId,
			paper_id: paperId,
			title: 'New Chat',
			created_at: '2026-01-01T00:00:00Z'
		});
		vi.mocked(sendMessage).mockImplementation(
			async (_p, _c, _content, _onDelta, onDone) => { onDone(); }
		);

		await createSession(paperId);
		await sendChatMessage(paperId, getActiveSessionId()!, 'Hello');

		const msgs = getMessages();
		const userMsg = msgs.find(m => m.role === 'user');
		expect(userMsg!.chat_session_id).toBe(realChatId);
	});
});

describe('draft discard on session switch', () => {
	const realChatId = 'existing-chat-id';

	it('removes draft from sessions when switching to a real session', async () => {
		vi.mocked(getChatSession).mockResolvedValue({
			id: realChatId,
			paper_id: paperId,
			title: 'Existing Chat',
			created_at: '2026-01-01T00:00:00Z',
			messages: []
		});
		vi.mocked(listChatSessions).mockResolvedValue([{
			id: realChatId,
			paper_id: paperId,
			title: 'Existing Chat',
			created_at: '2026-01-01T00:00:00Z'
		}]);

		// Load real sessions, then create a draft
		await loadSessions(paperId);
		await createSession(paperId);
		expect(getSessions()).toHaveLength(2);
		expect(getSessions().find(s => s.id.startsWith('draft-'))).toBeDefined();

		// Switch to the real session — draft should be discarded
		await selectSession(paperId, realChatId);

		expect(getSessions()).toHaveLength(1);
		expect(getSessions().find(s => s.id.startsWith('draft-'))).toBeUndefined();
		expect(getActiveSessionId()).toBe(realChatId);
	});

	it('removes draft when loadSessions is called (e.g. paper switch)', async () => {
		vi.mocked(listChatSessions).mockResolvedValue([]);

		await createSession(paperId);
		expect(getSessions()).toHaveLength(1);

		await loadSessions(paperId);
		expect(getSessions().find(s => s.id.startsWith('draft-'))).toBeUndefined();
	});
});

describe('draft deletion', () => {
	it('deletes draft locally without calling the API', async () => {
		await createSession(paperId);
		const draftId = getActiveSessionId()!;
		expect(draftId).toMatch(/^draft-/);

		await deleteSession(paperId, draftId);

		expect(deleteChatSession).not.toHaveBeenCalled();
		expect(getSessions()).toHaveLength(0);
		expect(getActiveSessionId()).toBeNull();
	});
});
