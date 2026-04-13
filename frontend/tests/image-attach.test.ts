import { describe, it, expect, beforeEach } from 'vitest';
import { addAttachment, consumeAttachments, getPendingAttachments, removeAttachment } from '../src/lib/attachments.svelte';

beforeEach(() => {
	// drain any leftover attachments from prior tests
	consumeAttachments();
});

describe('paste image → attachment strip → send flow', () => {
	it('pasted image appears in pending attachments with page 0', () => {
		addAttachment({ image_data: 'AAAA', text: '', page: 0 });
		const pending = getPendingAttachments();
		expect(pending).toHaveLength(1);
		expect(pending[0].image_data).toBe('AAAA');
		expect(pending[0].page).toBe(0);
		expect(pending[0].text).toBe('');
	});

	it('multiple images accumulate in the strip', () => {
		addAttachment({ image_data: 'IMG1', text: '', page: 0 });
		addAttachment({ image_data: 'IMG2', text: '', page: 0 });
		expect(getPendingAttachments()).toHaveLength(2);
	});

	it('consumeAttachments returns all and clears the strip', () => {
		addAttachment({ image_data: 'IMG1', text: '', page: 0 });
		addAttachment({ image_data: 'IMG2', text: '', page: 0 });
		const consumed = consumeAttachments();
		expect(consumed).toHaveLength(2);
		expect(consumed[0]).toEqual({ image_data: 'IMG1', text: '', page: 0 });
		expect(consumed[1]).toEqual({ image_data: 'IMG2', text: '', page: 0 });
		expect(getPendingAttachments()).toHaveLength(0);
	});

	it('removing an attachment by id works', () => {
		addAttachment({ image_data: 'IMG1', text: '', page: 0 });
		addAttachment({ image_data: 'IMG2', text: '', page: 0 });
		const id = getPendingAttachments()[0].id;
		removeAttachment(id);
		const remaining = getPendingAttachments();
		expect(remaining).toHaveLength(1);
		expect(remaining[0].image_data).toBe('IMG2');
	});

	it('send guard: attachments alone satisfy the send condition', () => {
		addAttachment({ image_data: 'AAAA', text: '', page: 0 });
		const content = '';
		const atts = consumeAttachments();
		const canSend = content.trim().length > 0 || atts.length > 0;
		expect(canSend).toBe(true);
	});

	it('send guard: empty text and no attachments blocks send', () => {
		const content = '';
		const atts = consumeAttachments();
		const canSend = content.trim().length > 0 || atts.length > 0;
		expect(canSend).toBe(false);
	});
});

describe('clipboard image extraction', () => {
	it('extracts image file from paste event items', () => {
		const file = new File(['pixels'], 'screenshot.png', { type: 'image/png' });
		const item = {
			type: 'image/png',
			getAsFile: () => file,
		} as unknown as DataTransferItem;

		const items = [item];
		const imageItem = items.find(i => i.type.startsWith('image/'));
		expect(imageItem).toBeDefined();
		expect(imageItem!.getAsFile()).toBe(file);
	});

	it('ignores non-image clipboard items', () => {
		const item = {
			type: 'text/plain',
			getAsFile: () => null,
		} as unknown as DataTransferItem;

		const items = [item];
		const imageItem = items.find(i => i.type.startsWith('image/'));
		expect(imageItem).toBeUndefined();
	});
});

describe('drop event image extraction', () => {
	it('filters image files from dropped file list', () => {
		const img = new File(['px'], 'photo.jpg', { type: 'image/jpeg' });
		const txt = new File(['hi'], 'notes.txt', { type: 'text/plain' });
		const files = [img, txt];

		const images = files.filter(f => f.type.startsWith('image/'));
		expect(images).toHaveLength(1);
		expect(images[0]).toBe(img);
	});
});
