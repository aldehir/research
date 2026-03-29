import type { MessageAttachment } from '$lib/api';

export interface PendingAttachment extends MessageAttachment {
	id: string;
}

let pending = $state<PendingAttachment[]>([]);

let nextId = 0;

export function addAttachment(att: MessageAttachment): void {
	pending = [...pending, { ...att, id: `att-${nextId++}` }];
}

export function removeAttachment(id: string): void {
	pending = pending.filter(a => a.id !== id);
}

export function consumeAttachments(): MessageAttachment[] {
	const result = pending.map(({ image_data, text, page }) => ({ image_data, text, page }));
	pending = [];
	return result;
}

export function getPendingAttachments(): PendingAttachment[] {
	return pending;
}
