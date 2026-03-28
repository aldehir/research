// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
	setSelectedText,
	getSelectedText,
	clearSelectedText,
	setPages,
	getSurroundingText,
	getCurrentPage,
	setCurrentPage
} from '$lib/pdf-context.svelte';

function makeFakePage(text: string) {
	return {
		getTextContent: async () => ({
			items: [{ str: text, hasEOL: false, dir: 'ltr', width: 100, height: 12, transform: [12, 0, 0, 12, 0, 0], fontName: 'f1' }],
			styles: {}
		})
	};
}

describe('pdf-context', () => {
	beforeEach(() => {
		clearSelectedText();
		setPages([]);
		setCurrentPage(1);
	});

	describe('selected text', () => {
		it('stores and retrieves selected text', () => {
			setSelectedText('hello world');
			expect(getSelectedText()).toBe('hello world');
		});

		it('clears selected text', () => {
			setSelectedText('something');
			clearSelectedText();
			expect(getSelectedText()).toBe('');
		});

		it('starts with empty selected text', () => {
			expect(getSelectedText()).toBe('');
		});
	});

	describe('surrounding text', () => {
		it('returns text from prev, current, and next pages', async () => {
			const pages = [
				makeFakePage('Page 1'),
				makeFakePage('Page 2'),
				makeFakePage('Page 3'),
				makeFakePage('Page 4')
			];
			setPages(pages as any);
			setCurrentPage(2); // 1-indexed

			const text = await getSurroundingText();
			expect(text).toContain('Page 1');
			expect(text).toContain('Page 2');
			expect(text).toContain('Page 3');
			expect(text).not.toContain('Page 4');
		});

		it('clamps to first page when on page 1', async () => {
			const pages = [
				makeFakePage('First'),
				makeFakePage('Second')
			];
			setPages(pages as any);
			setCurrentPage(1);

			const text = await getSurroundingText();
			expect(text).toContain('First');
			expect(text).toContain('Second');
		});

		it('clamps to last page when on last page', async () => {
			const pages = [
				makeFakePage('First'),
				makeFakePage('Second'),
				makeFakePage('Third')
			];
			setPages(pages as any);
			setCurrentPage(3);

			const text = await getSurroundingText();
			expect(text).toContain('Second');
			expect(text).toContain('Third');
			expect(text).not.toContain('First');
		});

		it('returns empty string when no pages are loaded', async () => {
			const text = await getSurroundingText();
			expect(text).toBe('');
		});
	});
});
