import {
	listChatSessions,
	createChatSession,
	getChatSession,
	deleteChatSession,
	sendMessage
} from '$lib/api';
import type { ChatSession, Message, MessageContext, ToolCall } from '$lib/api';
import { generateId } from '$lib/uuid';
import { requestGoToPage } from '$lib/pdf-navigate.svelte';

let sessions = $state<ChatSession[]>([]);
let activeSessionId = $state<string | null>(null);
let messages = $state<Message[]>([]);
let streamingContent = $state('');
let isStreaming = $state(false);
let activeToolCall = $state<ToolCall | null>(null);
let toolCallHandler = $state<((tool: ToolCall) => void) | null>(null);

export async function loadSessions(paperId: string): Promise<void> {
	sessions = await listChatSessions(paperId);
	if (activeSessionId && !sessions.find(s => s.id === activeSessionId)) {
		activeSessionId = null;
		messages = [];
	}
}

export async function createSession(paperId: string): Promise<void> {
	const session = await createChatSession(paperId);
	sessions = [session, ...sessions];
	activeSessionId = session.id;
	messages = [];
}

export async function selectSession(paperId: string, chatId: string): Promise<void> {
	activeSessionId = chatId;
	const session = await getChatSession(paperId, chatId);
	messages = session.messages;
}

export async function deleteSession(paperId: string, chatId: string): Promise<void> {
	await deleteChatSession(paperId, chatId);
	sessions = sessions.filter(s => s.id !== chatId);
	if (activeSessionId === chatId) {
		activeSessionId = null;
		messages = [];
	}
}

export async function sendChatMessage(
	paperId: string,
	chatId: string,
	content: string,
	context?: MessageContext
): Promise<void> {
	const userMessage: Message = {
		id: generateId(),
		chat_session_id: chatId,
		role: 'user',
		content,
		created_at: new Date().toISOString()
	};
	messages = [...messages, userMessage];
	isStreaming = true;
	streamingContent = '';

	await sendMessage(
		paperId,
		chatId,
		content,
		(text: string) => {
			streamingContent += text;
		},
		() => {
			const assistantMessage: Message = {
				id: generateId(),
				chat_session_id: chatId,
				role: 'assistant',
				content: streamingContent,
				created_at: new Date().toISOString()
			};
			messages = [...messages, assistantMessage];
			streamingContent = '';
			isStreaming = false;
			activeToolCall = null;
		},
		(error: string) => {
			console.error('Chat error:', error);
			streamingContent = '';
			isStreaming = false;
			activeToolCall = null;
		},
		context,
		(tool: ToolCall) => {
			activeToolCall = tool;
			if (tool.name === 'go_to_page' && typeof tool.args.page === 'number') {
				requestGoToPage(tool.args.page);
			}
			if (toolCallHandler) {
				toolCallHandler(tool);
			}
		}
	);
}

export function resetChat(): void {
	sessions = [];
	activeSessionId = null;
	messages = [];
	streamingContent = '';
	isStreaming = false;
}

export function getSessions(): ChatSession[] {
	return sessions;
}

export function getActiveSessionId(): string | null {
	return activeSessionId;
}

export function getMessages(): Message[] {
	return messages;
}

export function getStreamingContent(): string {
	return streamingContent;
}

export function getIsStreaming(): boolean {
	return isStreaming;
}

export function getActiveToolCall(): ToolCall | null {
	return activeToolCall;
}

export function setToolCallHandler(handler: ((tool: ToolCall) => void) | null): void {
	toolCallHandler = handler;
}
