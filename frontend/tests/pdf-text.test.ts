import { describe, it, expect } from 'vitest';
import { extractPageText } from '$lib/pdf-text';

function makeFakePage(items: { str: string; hasEOL: boolean }[]) {
	return {
		getTextContent: async () => ({
			items: items.map(item => ({
				str: item.str,
				hasEOL: item.hasEOL,
				dir: 'ltr',
				width: 100,
				height: 12,
				transform: [12, 0, 0, 12, 0, 0],
				fontName: 'g_d0_f1'
			})),
			styles: {}
		})
	};
}

describe('extractPageText', () => {
	it('concatenates text items into a string', async () => {
		const page = makeFakePage([
			{ str: 'Hello ', hasEOL: false },
			{ str: 'world', hasEOL: false }
		]);
		const text = await extractPageText(page as any);
		expect(text).toBe('Hello world');
	});

	it('adds newlines for items with hasEOL', async () => {
		const page = makeFakePage([
			{ str: 'Line one', hasEOL: true },
			{ str: 'Line two', hasEOL: false }
		]);
		const text = await extractPageText(page as any);
		expect(text).toBe('Line one\nLine two');
	});

	it('returns empty string for page with no text', async () => {
		const page = makeFakePage([]);
		const text = await extractPageText(page as any);
		expect(text).toBe('');
	});

	it('handles multiple EOL items', async () => {
		const page = makeFakePage([
			{ str: 'A', hasEOL: true },
			{ str: 'B', hasEOL: true },
			{ str: 'C', hasEOL: false }
		]);
		const text = await extractPageText(page as any);
		expect(text).toBe('A\nB\nC');
	});
});
