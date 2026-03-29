import {
	listChatSessions,
	createChatSession,
	getChatSession,
	deleteChatSession,
	sendMessage
} from '$lib/api';
import type { ChatSession, Message, ToolCall, ToolResult, MessageAttachment } from '$lib/api';
import { generateId } from '$lib/uuid';
import { requestGoToPage } from '$lib/pdf-navigate.svelte';

export type StreamSegment =
	| { type: 'text'; content: string }
	| { type: 'tool'; name: string; args: Record<string, unknown>; result?: ToolResult };

let sessions = $state<ChatSession[]>([]);
let activeSessionId = $state<string | null>(null);
let messages = $state<Message[]>([]);
let streamingContent = $state('');
let streamSegments = $state<StreamSegment[]>([]);
let messageSegments = $state(new Map<string, StreamSegment[]>());
let messageAttachments = $state(new Map<string, MessageAttachment[]>());
let isStreaming = $state(false);
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

function appendTextDelta(text: string): void {
	const last = streamSegments[streamSegments.length - 1];
	if (last?.type === 'text') {
		// Direct proxy mutation — Svelte 5 tracks the property write
		last.content += text;
	} else {
		streamSegments.push({ type: 'text', content: text });
	}
}

function snapshotSegments(): StreamSegment[] {
	return streamSegments.map(s =>
		s.type === 'text'
			? { type: 'text' as const, content: s.content }
			: { type: 'tool' as const, name: s.name, args: { ...s.args }, result: s.result }
	);
}

export async function sendChatMessage(
	paperId: string,
	chatId: string,
	content: string,
	currentPage?: number,
	attachments?: MessageAttachment[]
): Promise<void> {
	const userMessage: Message = {
		id: generateId(),
		chat_session_id: chatId,
		role: 'user',
		content,
		created_at: new Date().toISOString()
	};
	messages = [...messages, userMessage];
	if (attachments && attachments.length > 0) {
		messageAttachments = new Map(messageAttachments).set(userMessage.id, attachments);
	}
	isStreaming = true;
	streamingContent = '';
	streamSegments = [];

	await sendMessage(
		paperId,
		chatId,
		content,
		(text: string) => {
			streamingContent += text;
			appendTextDelta(text);
		},
		() => {
			const assistantMessage: Message = {
				id: generateId(),
				chat_session_id: chatId,
				role: 'assistant',
				content: streamingContent,
				created_at: new Date().toISOString()
			};
			const hasToolSegments = streamSegments.some(s => s.type === 'tool');
			if (hasToolSegments) {
				messageSegments = new Map(messageSegments).set(assistantMessage.id, snapshotSegments());
			}
			messages = [...messages, assistantMessage];
			streamingContent = '';
			streamSegments = [];
			isStreaming = false;
		},
		(error: string) => {
			console.error('Chat error:', error);
			streamingContent = '';
			streamSegments = [];
			isStreaming = false;
		},
		currentPage,
		(tool: ToolCall) => {
			streamSegments.push({
				type: 'tool',
				name: tool.name,
				args: tool.args
			});
			if (tool.name === 'go_to_page' && typeof tool.args.page === 'number') {
				requestGoToPage(tool.args.page);
			}
			if (toolCallHandler) {
				toolCallHandler(tool);
			}
		},
		(result: ToolResult) => {
			for (let i = streamSegments.length - 1; i >= 0; i--) {
				const seg = streamSegments[i];
				if (seg.type === 'tool' && seg.name === result.name && !seg.result) {
					seg.result = result;
					break;
				}
			}
		},
		attachments
	);
}

export function resetChat(): void {
	sessions = [];
	activeSessionId = null;
	messages = [];
	streamingContent = '';
	streamSegments = [];
	messageSegments = new Map();
	messageAttachments = new Map();
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

export function getStreamSegments(): StreamSegment[] {
	return streamSegments;
}

export function getMessageSegments(messageId: string): StreamSegment[] | undefined {
	return messageSegments.get(messageId);
}

export function getUserAttachments(messageId: string): MessageAttachment[] | undefined {
	return messageAttachments.get(messageId);
}

export function setToolCallHandler(handler: ((tool: ToolCall) => void) | null): void {
	toolCallHandler = handler;
}
