export interface Paper {
	id: string;
	title: string;
	file_path: string;
	file_size: number;
	created_at: string;
}

async function handleResponse<T>(response: Response): Promise<T> {
	if (!response.ok) {
		const body = await response.json() as { error: string };
		throw new Error(body.error);
	}
	return response.json() as Promise<T>;
}

export async function listPapers(): Promise<Paper[]> {
	const response = await fetch('/api/papers');
	return handleResponse<Paper[]>(response);
}

export async function uploadPaper(file: File): Promise<Paper> {
	const formData = new FormData();
	formData.append('file', file);
	const response = await fetch('/api/papers', {
		method: 'POST',
		body: formData
	});
	return handleResponse<Paper>(response);
}

export async function getPaper(id: string): Promise<Paper> {
	const response = await fetch(`/api/papers/${id}`);
	return handleResponse<Paper>(response);
}

export async function deletePaper(id: string): Promise<void> {
	const response = await fetch(`/api/papers/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		const body = await response.json() as { error: string };
		throw new Error(body.error);
	}
}

export function getPdfUrl(id: string): string {
	return `/api/papers/${id}/pdf`;
}

export interface ChatSession {
	id: string;
	paper_id: string;
	title: string;
	created_at: string;
}

export interface Message {
	id: string;
	chat_session_id: string;
	role: 'user' | 'assistant';
	content: string;
	selected_text?: string;
	created_at: string;
}

export interface MessageContext {
	selectedText?: string;
	surroundingText?: string;
	currentPage?: number;
}

export interface ChatSessionWithMessages extends ChatSession {
	messages: Message[];
}

export async function listChatSessions(paperId: string): Promise<ChatSession[]> {
	const response = await fetch(`/api/papers/${paperId}/chats`);
	return handleResponse<ChatSession[]>(response);
}

export async function createChatSession(paperId: string, title?: string): Promise<ChatSession> {
	const body: Record<string, string> = {};
	if (title !== undefined) {
		body.title = title;
	}
	const response = await fetch(`/api/papers/${paperId}/chats`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	return handleResponse<ChatSession>(response);
}

export async function getChatSession(paperId: string, chatId: string): Promise<ChatSessionWithMessages> {
	const response = await fetch(`/api/papers/${paperId}/chats/${chatId}`);
	return handleResponse<ChatSessionWithMessages>(response);
}

export async function deleteChatSession(paperId: string, chatId: string): Promise<void> {
	const response = await fetch(`/api/papers/${paperId}/chats/${chatId}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		const body = await response.json() as { error: string };
		throw new Error(body.error);
	}
}

export interface ToolCall {
	name: string;
	args: Record<string, unknown>;
}

export async function sendMessage(
	paperId: string,
	chatId: string,
	content: string,
	onDelta: (text: string) => void,
	onDone: () => void,
	onError: (error: string) => void,
	context?: MessageContext,
	onToolCall?: (tool: ToolCall) => void
): Promise<void> {
	const reqBody: Record<string, string | number> = { content };
	if (context?.selectedText) {
		reqBody.selected_text = context.selectedText;
	}
	if (context?.surroundingText) {
		reqBody.surrounding_text = context.surroundingText;
	}
	if (context?.currentPage) {
		reqBody.current_page = context.currentPage;
	}

	let response: Response;
	try {
		response = await fetch(`/api/papers/${paperId}/chats/${chatId}/messages`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(reqBody)
		});
	} catch (err) {
		onError(err instanceof Error ? err.message : 'Network error');
		return;
	}

	if (!response.ok) {
		const body = await response.json() as { error: string };
		onError(body.error);
		return;
	}

	const reader = response.body!.getReader();
	const decoder = new TextDecoder();
	let buffer = '';

	try {
		while (true) {
			const { done, value } = await reader.read();
			if (done) break;

			buffer += decoder.decode(value, { stream: true });
			const lines = buffer.split('\n');
			buffer = lines.pop() ?? '';

			for (const line of lines) {
				if (!line.startsWith('data: ')) continue;
				const json = line.slice(6);
				const event = JSON.parse(json) as { type: string; text?: string; name?: string; args?: Record<string, unknown> };
				if (event.type === 'delta' && event.text) {
					onDelta(event.text);
				} else if (event.type === 'tool_call' && onToolCall && event.name) {
					onToolCall({ name: event.name, args: event.args ?? {} });
				} else if (event.type === 'done') {
					onDone();
					return;
				}
			}
		}
		// Process remaining buffer
		if (buffer.startsWith('data: ')) {
			const json = buffer.slice(6);
			const event = JSON.parse(json) as { type: string; text?: string };
			if (event.type === 'delta' && event.text) {
				onDelta(event.text);
			}
			if (event.type === 'done') {
				onDone();
				return;
			}
		}
		onDone();
	} catch (err) {
		onError(err instanceof Error ? err.message : 'Stream error');
	}
}
