import { describe, it, expect, vi } from 'vitest';
import { extractOutline, type TocEntry } from '$lib/pdf-outline';

function makeDoc(outline: any[] | null, destinations?: Record<string, any>) {
	return {
		getOutline: vi.fn(async () => outline),
		getDestination: vi.fn(async (name: string) => destinations?.[name] ?? null),
		getPageIndex: vi.fn(async (ref: any) => ref.__pageIndex ?? 0)
	};
}

function ref(pageIndex: number) {
	return { __pageIndex: pageIndex };
}

describe('extractOutline', () => {
	it('returns empty array when outline is null', async () => {
		const doc = makeDoc(null);
		const result = await extractOutline(doc as any);
		expect(result).toEqual([]);
	});

	it('returns empty array when outline is empty', async () => {
		const doc = makeDoc([]);
		const result = await extractOutline(doc as any);
		expect(result).toEqual([]);
	});

	it('extracts flat outline with explicit destinations', async () => {
		const outline = [
			{ title: 'Chapter 1', dest: [ref(0), { name: 'Fit' }], items: [] },
			{ title: 'Chapter 2', dest: [ref(4), { name: 'Fit' }], items: [] }
		];
		const doc = makeDoc(outline);
		const result = await extractOutline(doc as any);

		expect(result).toEqual([
			{ title: 'Chapter 1', pageNumber: 1, children: [] },
			{ title: 'Chapter 2', pageNumber: 5, children: [] }
		]);
	});

	it('extracts outline with named string destinations', async () => {
		const outline = [
			{ title: 'Intro', dest: 'intro-dest', items: [] }
		];
		const doc = makeDoc(outline, {
			'intro-dest': [ref(2), { name: 'Fit' }]
		});
		const result = await extractOutline(doc as any);

		expect(result).toEqual([
			{ title: 'Intro', pageNumber: 3, children: [] }
		]);
	});

	it('extracts nested outline recursively', async () => {
		const outline = [
			{
				title: 'Part 1',
				dest: [ref(0), { name: 'Fit' }],
				items: [
					{ title: 'Section 1.1', dest: [ref(1), { name: 'Fit' }], items: [] },
					{
						title: 'Section 1.2',
						dest: [ref(3), { name: 'Fit' }],
						items: [
							{ title: 'Sub 1.2.1', dest: [ref(5), { name: 'Fit' }], items: [] }
						]
					}
				]
			}
		];
		const doc = makeDoc(outline);
		const result = await extractOutline(doc as any);

		expect(result).toHaveLength(1);
		expect(result[0].title).toBe('Part 1');
		expect(result[0].children).toHaveLength(2);
		expect(result[0].children[1].children).toHaveLength(1);
		expect(result[0].children[1].children[0]).toEqual({
			title: 'Sub 1.2.1', pageNumber: 6, children: []
		});
	});

	it('handles entries with null dest gracefully', async () => {
		const outline = [
			{ title: 'No Dest', dest: null, items: [] },
			{ title: 'Has Dest', dest: [ref(1), { name: 'Fit' }], items: [] }
		];
		const doc = makeDoc(outline);
		const result = await extractOutline(doc as any);

		expect(result).toEqual([
			{ title: 'No Dest', pageNumber: 1, children: [] },
			{ title: 'Has Dest', pageNumber: 2, children: [] }
		]);
	});
});
